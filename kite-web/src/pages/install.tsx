import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { InstallKiteButton } from "@/components/common/PwaExperience";
import {
  AppleIcon,
  ArrowUpRightIcon,
  DownloadIcon,
  MonitorIcon,
  SmartphoneIcon,
} from "lucide-react";

const platforms = [
  {
    title: "Windows",
    icon: MonitorIcon,
    browsers: [
      {
        name: "Chrome",
        steps: [
          "Open kite.onl in Chrome and sign in.",
          "Select the install icon in the address bar, or open ⋮ > Cast, save, and share > Install page as app.",
          "Select Install. Kite opens in its own window and appears in your apps.",
        ],
      },
      {
        name: "Microsoft Edge",
        steps: [
          "Open kite.onl in Edge and sign in.",
          "Open Settings and more (⋯) > Apps > Install this site as an app.",
          "Confirm with Install. You can pin Kite to Start or the taskbar.",
        ],
      },
    ],
  },
  {
    title: "macOS",
    icon: AppleIcon,
    browsers: [
      {
        name: "Chrome",
        steps: [
          "Open kite.onl in Chrome and sign in.",
          "Select the install icon in the address bar, or open ⋮ > Cast, save, and share > Install page as app.",
          "Select Install. Find Kite in Launchpad or your Applications folder.",
        ],
      },
      {
        name: "Safari",
        steps: [
          "Open kite.onl in Safari and sign in.",
          "From the menu bar, choose File > Add to Dock.",
          "Choose Add. Kite is available from the Dock and Launchpad.",
        ],
      },
    ],
  },
  {
    title: "Linux",
    icon: MonitorIcon,
    browsers: [
      {
        name: "Chrome or Chromium",
        steps: [
          "Open kite.onl in Chrome or Chromium and sign in.",
          "Select the install icon in the address bar, or open ⋮ > Install Kite.",
          "Confirm the installation. Open Kite from your applications menu.",
        ],
      },
    ],
  },
  {
    title: "Android",
    icon: SmartphoneIcon,
    browsers: [
      {
        name: "Chrome",
        steps: [
          "Open kite.onl in Chrome and sign in.",
          "Tap ⋮ > Install app, or Add to Home screen if that is what Chrome shows.",
          "Confirm. Open Kite from your Home screen or app list.",
        ],
      },
    ],
  },
  {
    title: "iPhone and iPad",
    icon: SmartphoneIcon,
    browsers: [
      {
        name: "Safari",
        steps: [
          "Open kite.onl in Safari and sign in.",
          "Tap the Share button, then Add to Home Screen.",
          "If shown, turn on Open as Web App, then tap Add. Launch Kite from your Home screen.",
        ],
      },
    ],
  },
];

export default function InstallPage() {
  return (
    <HomeLayout
      title="Install Kite on your device"
      description="Install Kite from its website on Windows, Mac, Linux, Android, iPhone, or iPad."
    >
      <main className="container max-w-5xl pb-20">
        <section className="pt-20 md:pt-28">
          <div className="flex size-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <DownloadIcon className="size-6" />
          </div>
          <h1 className="mt-5 text-4xl font-bold md:text-5xl">
            Install Kite on your device
          </h1>
          <p className="mt-5 max-w-3xl text-lg text-muted-foreground">
            Add Kite from kite.onl and open it in its own app window.
            There&apos;s no app store, separate download, or account setup
            beyond your Kite sign-in.
          </p>
          <InstallKiteButton className="mt-7" />
          <p className="mt-3 text-sm text-muted-foreground">
            If your browser doesn&apos;t show an install button, follow the
            steps for your device below.
          </p>
        </section>

        <section className="mt-14 grid gap-5 md:grid-cols-2">
          {platforms.map(({ title, icon: Icon, browsers }) => (
            <article key={title} className="rounded-xl border bg-card p-6">
              <div className="flex items-center gap-3">
                <Icon className="size-5 text-primary" aria-hidden="true" />
                <h2 className="text-xl font-semibold">{title}</h2>
              </div>
              <div className="mt-5 space-y-6">
                {browsers.map((browser) => (
                  <div key={browser.name}>
                    <h3 className="font-medium">{browser.name}</h3>
                    <ol className="mt-3 space-y-2.5">
                      {browser.steps.map((step, index) => (
                        <li
                          key={step}
                          className="flex items-start gap-3 text-sm leading-relaxed text-muted-foreground"
                        >
                          <span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium text-foreground">
                            {index + 1}
                          </span>
                          <span>{step}</span>
                        </li>
                      ))}
                    </ol>
                  </div>
                ))}
              </div>
            </article>
          ))}
        </section>

        <section className="mt-10 rounded-xl border bg-muted/40 p-6 md:p-8">
          <div className="flex items-center gap-3">
            <ArrowUpRightIcon className="size-5 text-primary" />
            <h2 className="text-xl font-semibold">Updates come from Kite</h2>
          </div>
          <p className="mt-3 max-w-3xl text-muted-foreground">
            Kite checks for website updates as you use it. When an update is
            ready, you&apos;ll see an option to reload and apply it. Save any
            unsaved changes before reloading. You never need to reinstall Kite
            or wait for an app store release.
          </p>
          <p className="mt-3 max-w-3xl text-sm text-muted-foreground">
            Kite needs an internet connection to load pages and sync with your
            account. If your connection drops, the current page may remain open,
            but server features won&apos;t work until you reconnect.
          </p>
        </section>
      </main>
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
