import Link from "next/link";
import { Button } from "../ui/button";

export default function HomeCallToAction() {
  return (
    <div className="max-w-7xl mx-auto rounded-xl bg-secondary saturate-80 relative px-10 py-12 lg:px-20 flex items-center">
      <img
        src="/illustrations/firmware.svg"
        alt=""
        className="hidden lg:block absolute h-96 -right-10 -top-8"
      />

      <div className="max-w-lg xl:max-w-2xl">
        <div className="text-white text-3xl font-semibold mb-5">
          Let Kite take care of your Bot
        </div>
        <div className="text-gray-300 text-xl mb-10 leading-relaxed">
          Ready to experience the new way of building Discord Bots and benefit
          from the power of WebAssembly?
        </div>
        <div className="flex">
          <Button asChild size="lg">
            <Link href="/app">Join the Beta</Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
