from playwright.sync_api import sync_playwright

def run_cuj(page):
    page.goto("http://localhost:5173")
    page.wait_for_timeout(1000)

    # Navigate to Settings
    page.get_by_role("link", name="Settings", exact=True).click()
    page.wait_for_timeout(1000)

    # Change Log Level to DEBUG
    page.locator("#log_level").select_option("DEBUG")
    page.wait_for_timeout(500)

    # Save Strategy
    page.get_by_role("button", name="Save Strategy").click()
    page.wait_for_timeout(1000)

    # Navigate to Devices Tab
    page.get_by_role("button", name="Devices", exact=True).click()
    page.wait_for_timeout(1000)

    # Click Add Device Category: Inverter
    page.get_by_text("Inverter / Solar").click()
    page.wait_for_timeout(500)

    # Fill out form for Enerlution
    page.locator("#name").fill("Test Enerlution")
    page.locator("#template").select_option("enerlution_inverter")
    page.wait_for_timeout(1000)

    # Check if Hybrid Inverter Features appear
    # The checkbox labels are "Grid Meter Connected?" and "Battery Connected?"
    page.get_by_label("Grid Meter Connected?").check()
    page.wait_for_timeout(500)
    page.get_by_label("Battery Connected?").check()
    page.wait_for_timeout(500)

    # Fill out battery capacity
    page.locator("#battery_capacity").fill("10.5")
    page.wait_for_timeout(1000)

    page.get_by_role("button", name="Save Device").click()
    page.wait_for_timeout(1000)

    # Navigate to Logger
    page.get_by_role("link", name="Logger", exact=True).click()
    page.wait_for_timeout(1000)

    # Change log level filter to DEBUG
    page.locator("select").first.select_option("DEBUG")
    page.wait_for_timeout(1000)

    # Type in search query
    page.get_by_placeholder("Search logs...").fill("Strategy")
    page.wait_for_timeout(1000)

    page.screenshot(path="/home/jules/verification/screenshots/verification.png", full_page=True)
    page.wait_for_timeout(2000)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/home/jules/verification/videos",
            viewport={"width": 1280, "height": 720}
        )
        page = context.new_page()
        try:
            run_cuj(page)
        except Exception as e:
            print(f"Exception: {e}")
            page.screenshot(path="/home/jules/verification/screenshots/error.png", full_page=True)
        finally:
            context.close()
            browser.close()
