package com.lbs.fitnesstracker;

import android.app.AlarmManager;
import android.app.PendingIntent;
import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;

import java.util.Calendar;

final class ReminderScheduler {
    static final String ACTION_REMINDER = "com.lbs.fitnesstracker.action.REMINDER";

    private static final String PREFS = "fitness_reminder";
    private static final String KEY_ENABLED = "enabled";
    private static final String KEY_HOUR = "hour";
    private static final String KEY_MINUTE = "minute";
    private static final int DEFAULT_HOUR = 20;
    private static final int DEFAULT_MINUTE = 0;
    private static final int REQUEST_CODE = 19797;

    private ReminderScheduler() {
    }

    static boolean isEnabled(Context context) {
        return preferences(context).getBoolean(KEY_ENABLED, false);
    }

    static int getHour(Context context) {
        return preferences(context).getInt(KEY_HOUR, DEFAULT_HOUR);
    }

    static int getMinute(Context context) {
        return preferences(context).getInt(KEY_MINUTE, DEFAULT_MINUTE);
    }

    static void enable(Context context, int hour, int minute) {
        preferences(context).edit()
                .putBoolean(KEY_ENABLED, true)
                .putInt(KEY_HOUR, hour)
                .putInt(KEY_MINUTE, minute)
                .apply();
        schedule(context);
    }

    static void disable(Context context) {
        preferences(context).edit().putBoolean(KEY_ENABLED, false).apply();
        AlarmManager manager = (AlarmManager) context.getSystemService(Context.ALARM_SERVICE);
        if (manager != null) {
            manager.cancel(reminderIntent(context));
        }
    }

    static void schedule(Context context) {
        if (!isEnabled(context)) {
            return;
        }

        AlarmManager manager = (AlarmManager) context.getSystemService(Context.ALARM_SERVICE);
        if (manager == null) {
            return;
        }

        long triggerAt = nextTriggerMillis(getHour(context), getMinute(context), System.currentTimeMillis());
        manager.setInexactRepeating(
                AlarmManager.RTC_WAKEUP,
                triggerAt,
                AlarmManager.INTERVAL_DAY,
                reminderIntent(context));
    }

    static long nextTriggerMillis(int hour, int minute, long nowMillis) {
        Calendar next = Calendar.getInstance();
        next.setTimeInMillis(nowMillis);
        next.set(Calendar.HOUR_OF_DAY, hour);
        next.set(Calendar.MINUTE, minute);
        next.set(Calendar.SECOND, 0);
        next.set(Calendar.MILLISECOND, 0);
        if (next.getTimeInMillis() <= nowMillis) {
            next.add(Calendar.DAY_OF_YEAR, 1);
        }
        return next.getTimeInMillis();
    }

    private static SharedPreferences preferences(Context context) {
        return context.getSharedPreferences(PREFS, Context.MODE_PRIVATE);
    }

    private static PendingIntent reminderIntent(Context context) {
        Intent intent = new Intent(context, ReminderReceiver.class).setAction(ACTION_REMINDER);
        return PendingIntent.getBroadcast(
                context,
                REQUEST_CODE,
                intent,
                PendingIntent.FLAG_UPDATE_CURRENT | PendingIntent.FLAG_IMMUTABLE);
    }
}
