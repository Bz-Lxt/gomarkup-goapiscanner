import { test, expect } from '@playwright/test'

const web = process.env.WEB_BASE || 'http://localhost:28481'

test('console shell renders', async ({ page }) => {
  await page.goto(web)
  await expect(page.getByRole('heading', { name: '安全扫描控制台' })).toBeVisible()
  await expect(page.getByText('授权声明', { exact: false })).toBeVisible()
})
