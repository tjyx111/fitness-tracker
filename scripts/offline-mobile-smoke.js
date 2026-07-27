#!/usr/bin/env node

const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

const baseURL = process.env.FITNESS_UI_URL || 'http://127.0.0.1:18088';

function assert(condition, message) {
    if (!condition) {
        throw new Error(message);
    }
}

(async () => {
    const browser = await chromium.launch({ headless: true });
    const context = await browser.newContext({
        viewport: { width: 393, height: 852 },
        deviceScaleFactor: 1,
        isMobile: true,
        hasTouch: true,
    });
    const page = await context.newPage();
    await page.addInitScript(() => {
        window.FitnessAndroid = {
            openReminderSettings() {
                window.__fitnessReminderOpened = true;
            },
            reloadApp() {
            },
        };
    });

    try {
        await page.goto(baseURL, { waitUntil: 'networkidle' });
        await page.evaluate(() => navigator.serviceWorker.ready);
        await page.reload({ waitUntil: 'networkidle' });

        const reminderButton = page.locator('#android-reminder-btn');
        assert(await reminderButton.isVisible(), 'Android reminder button is not visible');
        await reminderButton.click();
        assert(await page.evaluate(() => window.__fitnessReminderOpened === true), 'Android reminder bridge was not called');

        await page.locator('.nav-btn[data-page="todos"]').click();
        await page.waitForTimeout(300);
        await page.locator('.nav-btn[data-page="training"]').click();

        await context.setOffline(true);
        await page.reload({ waitUntil: 'domcontentloaded' });
        await page.waitForTimeout(300);

        assert(await page.locator('#page-training').isVisible(), 'cached training page is not visible offline');
        assert(await page.locator('#connectivity-banner').isVisible(), 'offline status banner is not visible');
        assert(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1), 'offline page overflows horizontally');

        process.stdout.write('offline Android WebView smoke test passed\n');
    } finally {
        await context.setOffline(false);
        await browser.close();
    }
})().catch(error => {
    process.stderr.write(`${error.stack || error}\n`);
    process.exitCode = 1;
});
