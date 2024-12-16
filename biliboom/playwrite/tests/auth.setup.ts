import { test as setup, expect } from '@playwright/test';
import path from 'path';

const authFile = path.join(__dirname, '../auth.json');

setup('B 站登录，并保存 COOKIE', async ({ page }) => {
    // 跳转到 B 站
    await page.goto('https://www.bilibili.com/');

    // 点击登录按钮，出现扫码登录弹窗
    await page.getByText('登录', { exact: true }).click();

    // 等等扫码登陆，并在手机上点击确认
    await page.waitForResponse(async resp => {
        return resp.url().includes("passport.bilibili.com/x/passport-login/web/qrcode/poll") && await resp.json().then(res => res.code == 0 && res.data.code == 0)
    });

    // 保存 cookie 到文件
    await page.context().storageState({ path: authFile });
});