import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  fullyParallel: false,
  retries: 1,
  use: {
    baseURL: 'http://localhost:4173',
    ignoreHTTPSErrors: true,
    trace: 'off',
    launchOptions: {
      args: ['--disable-dev-shm-usage', '--no-sandbox', '--disable-gpu']
    }
  },
  webServer: {
    command: 'VITE_API_BASE=http://backend:8443/api/v1 vite --host 0.0.0.0 --port 4173',
    url: 'http://localhost:4173',
    reuseExistingServer: true,
    timeout: 120_000
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] }
    }
  ]
});
