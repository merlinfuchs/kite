export function getId() {
  const time = new Date().getTime();
  const random = Math.random() * 1000;
  return `${time}-${random}`;
}
