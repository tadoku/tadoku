// Evaluate on the contest registration page after selecting two languages.
(() => {
  const group = document.querySelector('.paper-combobox__chips').getBoundingClientRect();
  const trigger = document.querySelector('.paper-combobox__chips .paper-combobox__trigger').getBoundingClientRect();
  const result = {
    viewport: innerWidth,
    rightInset: group.right - trigger.right,
    target: { width: trigger.width, height: trigger.height },
    aligned: group.right - trigger.right < 8,
    usableTarget: trigger.width >= 44 && trigger.height >= 44,
  };
  if (!result.aligned || !result.usableTarget) throw new Error(JSON.stringify(result));
  return result;
})();
