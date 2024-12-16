import { test, expect } from '@playwright/test';
import fs from "fs"
function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}


test('获取个人主页视频列表', async ({ page }) => {
  let ID = 3546771443681290
  // 等待页面加载完毕
  await page.goto(`https://space.bilibili.com/${ID}/video`, { waitUntil: "commit" })

  // 获取视频总数
  let video_count = 0
  await page.waitForResponse(async resp => {
    if (!resp.url().includes("api.bilibili.com/x/space/navnum")) {
      return false
    }
    const res = await resp.json()
    if (res.code != 0) {
      return false
    }
    video_count = res.data.video
    return true
  }, { timeout: 1000 * 10 });

  console.log("视频总数", video_count)

  let videos: Object[] = []
  // 读取所有视频
  for (let page_number = 1; video_count > 0; page_number += 1) {
    // 自动翻页
    await page.goto(`https://space.bilibili.com/${ID}/video?tid=0&pn=${page_number}&keyword=&order=pubdate`, { waitUntil: "commit" })
    const resp = await page.waitForResponse(async resp => {
      return resp.url().includes("api.bilibili.com/x/space/wbi/arc/search")
    })
    const res = await resp.json()
    const page_videos: Object[] = res.data.list.vlist
    video_count -= page_videos.length
    console.log(`查询到 ${page_videos.length} 个，还剩 ${video_count} 个`)
    videos = videos.concat(page_videos)
    await sleep(5000);
  }
  fs.writeFileSync("test-results/video_list.json", JSON.stringify(videos))
});

test('TODO 获取未读评论列表', async ({ page }) => {
  await page.goto('https://www.bilibili.com/');
  await page.locator('.header-entry-mini').click();
  await page.locator('.header-entry-avatar').click();
  await page.getByRole('link', { name: '更多' }).first().click();
})
