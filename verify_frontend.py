from playwright.sync_api import sync_playwright

def run_cuj(page):
    # Navigate to the dashboard
    page.goto("http://localhost:8080")
    page.wait_for_timeout(1000)

    # 1. Take a screenshot of the dashboard to show the Battery select positioning
    page.screenshot(path="/home/jules/verification/screenshots/dashboard.png")
    page.wait_for_timeout(500)

    # Navigate to settings tab
    page.get_by_text("Settings").click()
    page.wait_for_timeout(1000)

    # 2. Take a screenshot of the settings tab to show the new Latitude and Longitude fields
    # We scroll down a bit to make sure they are visible
    page.evaluate("window.scrollBy(0, 500)")
    page.wait_for_timeout(1000)
    page.screenshot(path="/home/jules/verification/screenshots/settings.png")
    page.wait_for_timeout(1000)

if __name__ == "__main__":
    import os
    os.makedirs("/home/jules/verification/screenshots", exist_ok=True)
    os.makedirs("/home/jules/verification/videos", exist_ok=True)
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
