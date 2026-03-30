from playwright.sync_api import sync_playwright
import os

def run_cuj(page):
    # Navigate to the dashboard
    page.goto("http://localhost:8080")
    page.wait_for_timeout(500)

    # Navigate to Settings
    page.get_by_text("Settings").click()
    page.wait_for_timeout(1000)

    # Click on the System Info tab
    page.get_by_text("System Info").click()
    page.wait_for_timeout(1000)

    # Scroll down to show Build Number
    page.evaluate("window.scrollBy(0, 500)")
    page.wait_for_timeout(1000)

    # Take screenshot of System Info
    page.screenshot(path="/home/jules/verification/screenshots/system_info.png")
    page.wait_for_timeout(500)

    # Force navigate to Logger tab by finding the link with "Logger" text
    page.evaluate("Array.from(document.querySelectorAll('a')).find(e => e.textContent.includes('Logger')).click()")
    page.wait_for_timeout(1000)

    # Select log level WARN
    try:
        page.locator("select").select_option("WARN")
        page.wait_for_timeout(1000)
    except Exception as e:
        print(f"Select failed, grabbing screenshot of what is visible... {e}")

    # Take screenshot of Logger
    page.screenshot(path="/home/jules/verification/screenshots/logger.png")
    page.wait_for_timeout(1000)

if __name__ == "__main__":
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
