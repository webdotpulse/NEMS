from playwright.sync_api import sync_playwright

def run_cuj(page):
    page.goto("http://localhost:8080")
    page.wait_for_timeout(2000)

    # Click on the Settings tab (first occurrence is desktop nav, second is mobile, we can just use the desktop one)
    page.get_by_role("link", name="Settings").click()
    page.wait_for_timeout(1000)

    # Click on the System Info tab
    page.get_by_text("System Info").click()
    page.wait_for_timeout(1000)

    # The token field is pre-filled, so we should clear it if we want to type something else.
    # But wait, we added it to the Software Update section inside System Info! Let's scroll there or just take a screenshot
    # Fill out the GitHub Token field
    page.get_by_placeholder("GitHub Token (Private Repo)").fill("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")
    page.wait_for_timeout(1000)

    # Take screenshot at the key moment
    page.screenshot(path="/home/jules/verification/screenshots/verification2.png")
    page.wait_for_timeout(1000)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/home/jules/verification/videos",
            viewport={'width': 1280, 'height': 720}
        )
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()
