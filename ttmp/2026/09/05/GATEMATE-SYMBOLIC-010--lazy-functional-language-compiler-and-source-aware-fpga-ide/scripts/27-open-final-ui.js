async(page)=>{
 await page.goto('http://127.0.0.1:18092');
 await page.getByLabel('Example program').selectOption('squares');
 await page.getByRole('button',{name:'Compile',exact:true}).click();
 await page.getByText('✓ Checked and compiled',{exact:false}).waitFor();
 await page.getByLabel('Heap filter').selectOption('CONS');
 await page.locator('.heap-scroll tbody tr').first().click();
 return {url:page.url(),backend:await page.locator('.backend').innerText(),values:await page.locator('.stream-values>span').allTextContents()};
}
