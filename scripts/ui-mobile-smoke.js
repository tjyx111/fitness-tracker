#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

const baseURL = process.env.ASSISTANT_UI_URL || process.env.FITNESS_UI_URL || 'http://127.0.0.1:18088';
const outputDir = process.env.ASSISTANT_UI_OUTPUT || process.env.FITNESS_UI_OUTPUT || '/tmp/assistant-ui-smoke';
const pages = ['training', 'exercises', 'notes', 'todos', 'challenges', 'statistics', 'reports'];

function assert(condition, message) {
    if (!condition) {
        throw new Error(message);
    }
}

(async () => {
    fs.mkdirSync(outputDir, { recursive: true });
    const browser = await chromium.launch({ headless: true });
    const page = await browser.newPage({
        viewport: { width: 393, height: 852 },
        deviceScaleFactor: 1,
        isMobile: true,
        hasTouch: true,
    });

    try {
        await page.goto(baseURL, { waitUntil: 'networkidle' });

        const navButtons = page.locator('.nav-btn');
        assert(await navButtons.count() === pages.length, 'expected seven navigation buttons');

        for (let index = 0; index < pages.length; index += 1) {
            const pageName = pages[index];
            const button = navButtons.nth(index);
            const box = await button.boundingBox();
            assert(box && box.height >= 44, `${pageName} navigation target is shorter than 44px`);

            await button.click();
            await page.waitForTimeout(250);

            assert(await button.getAttribute('aria-selected') === 'true', `${pageName} tab is not selected`);
            assert(await page.locator(`#page-${pageName}`).evaluate((element) => element.classList.contains('active')), `${pageName} page is not active`);

            const layout = await page.evaluate(() => ({
                viewportWidth: window.innerWidth,
                documentWidth: document.documentElement.scrollWidth,
                navPosition: getComputedStyle(document.querySelector('.nav')).position,
            }));
            assert(layout.documentWidth <= layout.viewportWidth + 1, `${pageName} causes horizontal page overflow: ${layout.documentWidth}px > ${layout.viewportWidth}px`);
            assert(layout.navPosition === 'fixed', `${pageName} navigation is not fixed on mobile`);

            await page.screenshot({
                path: path.join(outputDir, `${pageName}.png`),
                fullPage: false,
            });

            if (pageName === 'training' && await page.locator('#groups-container .group-card').count() > 0) {
                await page.locator('#groups-container .group-card').first().click();
                await page.waitForTimeout(250);
                assert(await page.locator('#page-training-detail').evaluate((element) => element.classList.contains('active')), 'training detail page is not active');
                assert(await page.locator('.actions').evaluate((element) => getComputedStyle(element).position === 'sticky'), 'training actions are not sticky on mobile');
                const firstExercise = page.locator('.exercise-card').first();
                if (await firstExercise.count() > 0) {
                    const bottomAddButton = firstExercise.locator('.add-set-btn-bottom');
                    assert(await bottomAddButton.count() === 1, 'exercise has no bottom add-set button');
                    const initialSetCount = await firstExercise.locator('.current-set').count();
                    for (let click = 0; click < 8; click += 1) {
                        await bottomAddButton.click();
                    }
                    await page.waitForTimeout(100);
                    const finalSetCount = await firstExercise.locator('.current-set').count();
                    assert(finalSetCount === initialSetCount + 8, 'bottom add-set button did not append rows');
                    assert(await firstExercise.locator('.current-set').last().locator('.set-number-label').textContent() === `第${finalSetCount}组`, 'last set number is incorrect');
                    const lastSetBox = await firstExercise.locator('.current-set').last().boundingBox();
                    const bottomButtonBox = await bottomAddButton.boundingBox();
                    const stickyActionsBox = await page.locator('#page-training-detail .actions').boundingBox();
                    assert(lastSetBox && bottomButtonBox && bottomButtonBox.y >= lastSetBox.y + lastSetBox.height - 1, 'bottom add-set button is not after the final row');
                    assert(bottomButtonBox && stickyActionsBox && bottomButtonBox.y >= 0 && bottomButtonBox.y + bottomButtonBox.height <= stickyActionsBox.y - 8, 'bottom add-set button is hidden by sticky controls');
                    await page.screenshot({
                        path: path.join(outputDir, 'training-detail-long-sets.png'),
                        fullPage: false,
                    });
                }
                await page.screenshot({
                    path: path.join(outputDir, 'training-detail.png'),
                    fullPage: false,
                });
                await page.locator('#back-to-groups').click();
            }

            if (pageName === 'exercises' && await page.locator('#groups-table-container button[onclick^="editGroup"]').count() > 0) {
                await page.locator('#groups-table-container button[onclick^="editGroup"]').first().click();
                const selectedRows = page.locator('#exercise-checkboxes .group-exercise-row.selected');
                if (await selectedRows.count() >= 2) {
                    const firstID = await selectedRows.nth(0).getAttribute('data-exercise-id');
                    const secondID = await selectedRows.nth(1).getAttribute('data-exercise-id');
                    await selectedRows.nth(0).locator('.group-exercise-move-down').click();
                    assert(await selectedRows.nth(0).getAttribute('data-exercise-id') === secondID, 'group exercise did not move down');
                    assert(await selectedRows.nth(1).getAttribute('data-exercise-id') === firstID, 'group exercise order was not swapped');
                    await selectedRows.nth(1).locator('.group-exercise-move-up').click();
                }
                await page.locator('#group-modal .btn-secondary').first().click();
            }

            if (pageName === 'challenges' && await page.locator('.challenge-history-card').count() > 0) {
                await page.locator('.challenge-history-toggle').first().click();
                await page.locator('.challenge-history-day').first().waitFor();
                assert(await page.locator('.challenge-history-day').count() > 0, 'challenge history has no daily details');
                assert(await page.locator('.challenge-history-detail input[type="checkbox"]').count() === 0, 'challenge history is editable');
                assert(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1), 'challenge history causes horizontal overflow');
                await page.screenshot({
                    path: path.join(outputDir, 'challenges-history.png'),
                    fullPage: false,
                });
            }

            if (pageName === 'statistics' && await page.locator('.challenge-history-card').count() > 0) {
                await page.locator('#stats-type').selectOption('challenges');
                const trackedDay = page.locator('.heatmap-day[title*=" / "]').first();
                await trackedDay.waitFor();
                await trackedDay.click();
                await page.locator('#day-records .challenge-card').first().waitFor();
                assert(await page.locator('#day-records .challenge-status-badge').count() > 0, 'historical challenge day has no status');
                await page.screenshot({
                    path: path.join(outputDir, 'statistics-challenges.png'),
                    fullPage: false,
                });
            }

            if (pageName === 'reports' && await page.locator('.report-list-item').count() > 0) {
                const firstReport = page.locator('.report-list-item').first();
                await firstReport.click();
                const reportFrame = page.locator('#report-frame');
                assert(await reportFrame.isVisible(), 'selected HTML report is not visible');
                assert(Boolean(await reportFrame.getAttribute('src')), 'selected HTML report has no source URL');
            }

            await page.evaluate(() => window.scrollTo(0, document.documentElement.scrollHeight));
            assert(await button.isVisible(), `${pageName} navigation is not visible after scrolling`);
        }

        process.stdout.write(`mobile UI smoke test passed; screenshots: ${outputDir}\n`);
    } finally {
        await browser.close();
    }
})().catch((error) => {
    process.stderr.write(`${error.stack || error}\n`);
    process.exitCode = 1;
});
