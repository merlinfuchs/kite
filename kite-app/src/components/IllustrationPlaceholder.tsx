import clsx from "clsx";
import { useRouter } from "next/router";

interface Props {
  svgPath: string;
  title: string;
  className?: string;
}

export default function IllustrationPlaceholder({
  svgPath,
  title,
  className,
}: Props) {
  const router = useRouter();

  return (
    <div className={clsx("flex justify-center w-full", className)}>
      <div className="max-w-3xl max-h-96 h-full">
        <img
          src={router.basePath + svgPath}
          className="px-5 md:px-10 lg:px-32 w-full h-full"
          alt=""
        />
        <div className="text-center text-gray-400 font-light text-lg mt-10">
          {title}
        </div>
      </div>
    </div>
  );
}
