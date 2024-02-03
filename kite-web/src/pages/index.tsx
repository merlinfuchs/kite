import HomeLayout from "@/components/home/HomeLayout";
import HomeCodeShowcase from "@/components/home/HomeCodeShowcase";
import Link from "next/link";
import HomeFeatures from "@/components/home/HomeFeatures";
import HomeCallToAction from "@/components/home/HomeCallToAction";

export default function HomePage() {
  return (
    <HomeLayout>
      <div className="bg-dark-3 pb-32 pt-20 sm:pt-32 px-10 relative flex justify-center gap-20">
        <div className="max-w-[1400px] xl:w-full flex flex-col justify-between items-start xl:flex-row gap-16">
          <div className="text-center xl:pt-24 max-w-xl mx-auto">
            <h1 className="text-5xl text-white font-bold mb-10 leading-snug">
              The WebAssembly Runtime for Discord Bots
            </h1>
            <h2 className="text-gray-300 text-2xl font-base mb-10 leading-relaxed">
              Make Discord Bots without worrying about hosting and scaling.
              Concentrate on what you do best, building your bot.
            </h2>
            <div className="flex justify-center">
              <Link
                href="/app"
                className="bg-primary px-10 py-4 rounded-lg font-medium text-white text-xl block transition-transform hover:-translate-y-1"
              >
                Join the Beta!
              </Link>
            </div>
          </div>
          <div className="hidden sm:block">
            <HomeCodeShowcase />
          </div>
        </div>
      </div>
      <div className="bg-dark-4 pb-64">
        <HomeFeatures />
      </div>
      <div className="-mt-48 px-6 sm:px-10">
        <HomeCallToAction />
      </div>
    </HomeLayout>
  );
}
