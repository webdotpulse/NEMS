from playwright.sync_api import sync_playwright

def run_cuj(page):
    # Navigate to app
    page.goto("http://localhost:5173")
    page.wait_for_timeout(1000)

    # 1. Navigate to Settings
    page.get_by_role("link", name="Settings").click()
    page.wait_for_timeout(1000)

    # Take screenshot of Strategy / Grid config order
    page.get_by_role("button", name="Strategy").click()
    page.wait_for_timeout(500)
    page.screenshot(path="/home/jules/verification/screenshots/settings_strategy_order.png")

    # 2. Check Add Device Category Cards
    page.get_by_role("button", name="Devices").click()
    page.wait_for_timeout(1000)

    # Scroll to add device section
    page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
    page.wait_for_timeout(1000)

    # Take screenshot of category cards
    page.screenshot(path="/home/jules/verification/screenshots/add_device_categories.png")

    # Select category
    page.locator('div').filter(has_text='EV Charger').last.click()
    page.wait_for_timeout(1000)

    # Take screenshot of the filtered form
    page.screenshot(path="/home/jules/verification/screenshots/add_device_form.png")

    # 3. Check Network Scanner
    page.get_by_role("link", name="Scanner").click()
    page.wait_for_timeout(1000)

    page.get_by_role("button", name="Start Scan").click()
    page.wait_for_timeout(4000) # Give it some time to scan

    # Take screenshot of the scanner results (vendor column)
    page.screenshot(path="/home/jules/verification/screenshots/scanner_vendor.png")

    # 4. Dashboard Grid & Battery Nodes (Power Flow)
    page.get_by_role("link", name="Dashboard").click()
    page.wait_for_timeout(1000)
    page.screenshot(path="/home/jules/verification/screenshots/dashboard.png")

if __name__ == "__main__":
    import os
    os.makedirs("/home/jules/verification/videos", exist_ok=True)
    os.makedirs("/home/jules/verification/screenshots", exist_ok=True)
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/home/jules/verification/videos"
        )
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()
