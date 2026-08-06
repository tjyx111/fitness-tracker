# 助手 Android

This project builds an Android WebView application that loads the deployed Assistant at `https://111.230.63.109:19797/`.

The Android wrapper provides:

- app-private WebView and service-worker caches for the latest successfully loaded pages and GET API responses;
- a first-launch offline recovery screen;
- an Android-only reminder button with an optional daily notification;
- reminder restoration after device reboot;
- strict HTTPS certificate validation using the bundled private CA.

Offline mode is read-only. Creating or changing Assistant data still requires a network connection to the server.

Build the signed release APK from the repository root:

```bash
./scripts/build_android.sh
```

The release signing key and password are generated outside the repository under `/root/.config/fitness-tracker/android/`. Keep them for every future update: Android only permits an installed application to be upgraded by an APK signed with the same key.

Output:

```text
dist/assistant.apk
```
