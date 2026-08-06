package com.lbs.fitnesstracker;

import android.Manifest;
import android.annotation.SuppressLint;
import android.app.Activity;
import android.app.AlertDialog;
import android.content.Context;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.Uri;
import android.net.http.SslError;
import android.os.Build;
import android.os.Bundle;
import android.webkit.JavascriptInterface;
import android.webkit.ServiceWorkerController;
import android.webkit.SslErrorHandler;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.TimePicker;
import android.widget.Toast;

public final class MainActivity extends Activity {
    private static final String APP_URL = "https://111.230.63.109:19797/";
    private static final String APP_HOST = "111.230.63.109";
    private static final int APP_PORT = 19797;
    private static final int NOTIFICATION_PERMISSION_REQUEST = 1001;
    private static final String OFFLINE_HTML = "<!doctype html><html lang=\"zh-CN\"><head>"
            + "<meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\">"
            + "<style>body{margin:0;min-height:100vh;display:grid;place-items:center;background:#f4f7f6;"
            + "font-family:sans-serif;color:#17201f}.card{box-sizing:border-box;width:calc(100% - 32px);max-width:420px;"
            + "padding:28px 22px;text-align:center;background:white;border:1px solid #c8d6d2;border-radius:24px;"
            + "box-shadow:0 14px 36px rgba(15,65,59,.12)}.icon{font-size:42px}h1{font-size:22px;margin:14px 0 10px}"
            + "p{color:#5f6f6c;line-height:1.6;margin:0 0 22px}button{width:100%;min-height:50px;border:0;"
            + "border-radius:15px;background:#0f766e;color:white;font-size:16px;font-weight:700}</style></head><body>"
            + "<main class=\"card\"><div class=\"icon\">↻</div><h1>暂时无法连接助手</h1>"
            + "<p>首次使用需要联网。成功打开一次后，断网时仍可查看最近缓存的数据。</p>"
            + "<button onclick=\"AssistantAndroid.reloadApp()\">重新连接</button></main></body></html>";

    private WebView webView;
    private int pendingReminderHour = -1;
    private int pendingReminderMinute = -1;

    @SuppressLint({"SetJavaScriptEnabled", "AddJavascriptInterface"})
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        webView = findViewById(R.id.web_view);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setDatabaseEnabled(true);
        settings.setCacheMode(resolveCacheMode());
        settings.setAllowFileAccess(false);
        settings.setAllowContentAccess(false);
        settings.setAllowFileAccessFromFileURLs(false);
        settings.setAllowUniversalAccessFromFileURLs(false);
        settings.setMixedContentMode(WebSettings.MIXED_CONTENT_NEVER_ALLOW);
        settings.setSupportMultipleWindows(false);
        settings.setMediaPlaybackRequiresUserGesture(true);

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            ServiceWorkerController.getInstance()
                    .getServiceWorkerWebSettings()
                    .setCacheMode(WebSettings.LOAD_DEFAULT);
        }

        AndroidBridge androidBridge = new AndroidBridge();
        webView.addJavascriptInterface(androidBridge, "AssistantAndroid");
        webView.addJavascriptInterface(androidBridge, "FitnessAndroid");
        webView.setWebViewClient(new AssistantWebViewClient());
        ReminderScheduler.schedule(this);
        webView.loadUrl(APP_URL);
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (webView != null) {
            webView.getSettings().setCacheMode(resolveCacheMode());
        }
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
            return;
        }
        super.onBackPressed();
    }

    @Override
    protected void onDestroy() {
        if (webView != null) {
            webView.removeJavascriptInterface("AssistantAndroid");
            webView.removeJavascriptInterface("FitnessAndroid");
            webView.stopLoading();
            webView.destroy();
            webView = null;
        }
        super.onDestroy();
    }

    @Override
    public void onRequestPermissionsResult(
            int requestCode,
            String[] permissions,
            int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        if (requestCode != NOTIFICATION_PERMISSION_REQUEST) {
            return;
        }
        if (grantResults.length > 0 && grantResults[0] == PackageManager.PERMISSION_GRANTED
                && pendingReminderHour >= 0 && pendingReminderMinute >= 0) {
            enableReminder(pendingReminderHour, pendingReminderMinute);
        } else {
            ReminderScheduler.disable(this);
            Toast.makeText(this, R.string.reminder_permission_denied, Toast.LENGTH_LONG).show();
        }
        pendingReminderHour = -1;
        pendingReminderMinute = -1;
    }

    private int resolveCacheMode() {
        return isNetworkAvailable() ? WebSettings.LOAD_DEFAULT : WebSettings.LOAD_CACHE_ELSE_NETWORK;
    }

    private boolean isNetworkAvailable() {
        ConnectivityManager manager =
                (ConnectivityManager) getSystemService(Context.CONNECTIVITY_SERVICE);
        if (manager == null) {
            return false;
        }
        Network network = manager.getActiveNetwork();
        if (network == null) {
            return false;
        }
        NetworkCapabilities capabilities = manager.getNetworkCapabilities(network);
        return capabilities != null
                && capabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET);
    }

    private void reloadApp() {
        if (webView == null) {
            return;
        }
        webView.getSettings().setCacheMode(resolveCacheMode());
        webView.loadUrl(APP_URL);
    }

    private void openReportInBrowser(String rawUrl) {
        Uri uri = Uri.parse(rawUrl);
        int port = uri.getPort() == -1 ? 443 : uri.getPort();
        String path = uri.getPath();
        if (!"https".equalsIgnoreCase(uri.getScheme())
                || !APP_HOST.equalsIgnoreCase(uri.getHost())
                || port != APP_PORT
                || path == null
                || !path.startsWith("/api/reports/")) {
            Toast.makeText(this, R.string.report_link_invalid, Toast.LENGTH_LONG).show();
            return;
        }

        Intent intent = new Intent(Intent.ACTION_VIEW, uri);
        if (intent.resolveActivity(getPackageManager()) == null) {
            Toast.makeText(this, R.string.report_browser_unavailable, Toast.LENGTH_LONG).show();
            return;
        }
        startActivity(intent);
    }

    private void showReminderSettings() {
        TimePicker picker = new TimePicker(this);
        picker.setIs24HourView(true);
        picker.setHour(ReminderScheduler.getHour(this));
        picker.setMinute(ReminderScheduler.getMinute(this));

        String status = ReminderScheduler.isEnabled(this)
                ? getString(
                        R.string.reminder_enabled,
                        ReminderScheduler.getHour(this),
                        ReminderScheduler.getMinute(this))
                : getString(R.string.reminder_disabled);

        new AlertDialog.Builder(this)
                .setTitle(R.string.reminder_title)
                .setMessage(status)
                .setView(picker)
                .setPositiveButton(R.string.reminder_enable, (dialog, which) ->
                        requestReminderPermission(picker.getHour(), picker.getMinute()))
                .setNeutralButton(R.string.reminder_disable, (dialog, which) -> {
                    ReminderScheduler.disable(this);
                    Toast.makeText(this, R.string.reminder_disabled, Toast.LENGTH_SHORT).show();
                })
                .setNegativeButton(R.string.reminder_cancel, null)
                .show();
    }

    private void requestReminderPermission(int hour, int minute) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU
                && checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS)
                != PackageManager.PERMISSION_GRANTED) {
            pendingReminderHour = hour;
            pendingReminderMinute = minute;
            requestPermissions(
                    new String[]{Manifest.permission.POST_NOTIFICATIONS},
                    NOTIFICATION_PERMISSION_REQUEST);
            return;
        }
        enableReminder(hour, minute);
    }

    private void enableReminder(int hour, int minute) {
        ReminderScheduler.enable(this, hour, minute);
        Toast.makeText(
                this,
                getString(R.string.reminder_enabled, hour, minute),
                Toast.LENGTH_LONG).show();
    }

    private final class AndroidBridge {
        @JavascriptInterface
        public void openReminderSettings() {
            runOnUiThread(MainActivity.this::showReminderSettings);
        }

        @JavascriptInterface
        public void reloadApp() {
            runOnUiThread(MainActivity.this::reloadApp);
        }

        @JavascriptInterface
        public void openReportInBrowser(String url) {
            runOnUiThread(() -> MainActivity.this.openReportInBrowser(url));
        }
    }

    private final class AssistantWebViewClient extends WebViewClient {
        @Override
        public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
            Uri uri = request.getUrl();
            int port = uri.getPort() == -1 ? 443 : uri.getPort();
            if ("https".equalsIgnoreCase(uri.getScheme())
                    && APP_HOST.equalsIgnoreCase(uri.getHost())
                    && port == APP_PORT) {
                return false;
            }

            Intent intent = new Intent(Intent.ACTION_VIEW, uri);
            if (intent.resolveActivity(getPackageManager()) != null) {
                startActivity(intent);
            }
            return true;
        }

        @Override
        public void onReceivedError(
                WebView view,
                WebResourceRequest request,
                WebResourceError error) {
            if (request.isForMainFrame() && APP_HOST.equalsIgnoreCase(request.getUrl().getHost())) {
                view.loadDataWithBaseURL(null, OFFLINE_HTML, "text/html", "UTF-8", null);
            }
        }

        @Override
        public void onReceivedSslError(WebView view, SslErrorHandler handler, SslError error) {
            handler.cancel();
            Toast.makeText(MainActivity.this, R.string.tls_error, Toast.LENGTH_LONG).show();
        }
    }
}
