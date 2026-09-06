async(page)=>({text:await page.locator('body').innerText(),screenshots:await page.evaluate(()=>document.title)})
