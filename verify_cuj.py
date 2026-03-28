import time
from playwright.sync_api import sync_playwright

def run_cuj(page):
    # Navigate to the dashboard
    page.goto("http://localhost:5173")
    page.wait_for_timeout(2000)

    # Take screenshot of the power flow
    page.screenshot(path="/home/jules/verification/screenshots/dashboard.png")
    page.wait_for_timeout(1000)

    # Navigate to settings page
    page.get_by_text("Settings", exact=True).click()
    page.wait_for_timeout(1000)

    # Scroll down to Add Device
    page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
    page.wait_for_timeout(1000)

    # Take screenshot of the add device categories
    page.screenshot(path="/home/jules/verification/screenshots/settings.png")
    page.wait_for_timeout(1000)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/home/jules/verification/videos",
            viewport={"width": 1280, "height": 800}
        )
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()
