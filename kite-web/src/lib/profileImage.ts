// Profile images are sent as JSON with the image as a data URI, which the
// API's body limit has to fit.
export const maxProfileImageSize = 5 * 1024 * 1024;
export const profileImageTypes = "image/png,image/jpeg,image/gif,image/webp";

export function readDataUrl(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });
}
