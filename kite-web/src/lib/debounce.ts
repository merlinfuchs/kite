export default function debounce(func: () => void, timeout: number) {
  let timer: ReturnType<typeof setTimeout>;

  return () => {
    if (timer) {
      clearTimeout(timer);
    }
    timer = setTimeout(func, timeout);
  };
}
