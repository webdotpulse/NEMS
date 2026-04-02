from playwright.sync_api import sync_playwright

def run_cuj(page):
    page.goto("http://localhost:5173")  # Vite dev server
    page.wait_for_timeout(1000)

    # Click on the Settings tab (using exact role to avoid ambiguity)
    page.get_by_role("link", name="Settings").click()
    page.wait_for_timeout(1000)

    # Click on the System Info tab within Settings (it's a button)
    page.get_by_role("button", name="System Info").click()
    page.wait_for_timeout(1000)

    # Take screenshot of the System Info tab where the input was
    page.screenshot(path="/home/jules/verification/screenshots/verification.png", full_page=True)
    page.wait_for_timeout(1000)

if __name__ == "__main__":
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
