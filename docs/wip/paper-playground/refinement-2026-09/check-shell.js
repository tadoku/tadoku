// Evaluate in the collaborative browser at each review viewport.
// This checks the shared frame, independently of intentional nested columns.
(() => {
  const main = document.querySelector('.app-content');
  const brand = document.querySelector('.app-wordmark').getBoundingClientRect();
  const footer = document.querySelector('.app-footer__inner').getBoundingClientRect();
  const bounds = main.getBoundingClientRect();
  const left = bounds.left + parseFloat(getComputedStyle(main).paddingLeft);
  const right = bounds.right - parseFloat(getComputedStyle(main).paddingRight);
  const result = {
    viewport: innerWidth,
    content: { left, right },
    brandLeft: brand.left,
    footer: { left: footer.left, right: footer.right },
    aligned: Math.abs(brand.left - left) < 1 && Math.abs(footer.left - left) < 1 && Math.abs(footer.right - right) < 1,
    noPageOverflow: document.documentElement.scrollWidth <= innerWidth,
  };
  if (!result.aligned || !result.noPageOverflow) throw new Error(JSON.stringify(result));
  return result;
})();
