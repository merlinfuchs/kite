import { toast } from "sonner";

// Profile images are sent as JSON with the image as a data URI, which the
// API's body limit has to fit.
const maxProfileImageSize = 5 * 1024 * 1024;
export const profileImageTypes = "image/png,image/jpeg,image/gif,image/webp";

// readProfileImage returns the file as a data URI, or undefined with a toast
// if it's too large.
export function readProfileImage(file: File) {
  if (file.size > maxProfileImageSize) {
    toast.error("Images can be at most 5 MB");
    return Promise.resolve(undefined);
  }
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });
}
