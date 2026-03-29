from playwright.sync_api import sync_playwright

def run_cuj(page):
    # Navigate to the dashboard
    page.goto("http://localhost:8080")
    page.wait_for_timeout(1000)

    # We want to wait for the battery node to show up if it hasn't
    page.wait_for_timeout(2000)

    # Now that we have a battery, there's a `<select>` for the battery mode.
    # We can screenshot the dashboard again.
    # Oh, wait... is the battery showing up in the flow diagram? Let me see.
    page.evaluate("window.scrollBy(0, 300)")
    page.wait_for_timeout(1000)
    page.screenshot(path="/home/jules/verification/screenshots/dashboard_with_battery_flow.png")
    page.wait_for_timeout(500)

if __name__ == "__main__":
    import os
    os.makedirs("/home/jules/verification/screenshots", exist_ok=True)
    os.makedirs("/home/jules/verification/videos", exist_ok=True)
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/home/jules/verification/videos",
            viewport={"width": 1280, "height": 900}
        )
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()
