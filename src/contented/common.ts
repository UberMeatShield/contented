export function getWindowSizes() {
  let width = !(window as any).jasmine ? window.innerWidth : 800;
  let height = !(window as any).jasmine ? window.innerHeight : 800;
  return { width, height };
}
