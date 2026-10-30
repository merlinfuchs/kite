import { CheckIcon } from "lucide-react";
import Link from "next/link";
import {
  PersonaMockup,
  ProfileMockup,
  rotatingStatuses,
} from "./ProfileMockups";

const points = [
  "Set your own status and activity for free, like Watching #support.",
  "Rotate through statuses and give your bot an animated avatar and banner with Plus.",
  "Give it a profile per server and send messages with another name and avatar, like the user who ran a command, with Premium.",
];

export default function HomeProfileSection() {
  return (
    <section id="profile" className="scroll-mt-16 container py-16 sm:py-24">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
        <div className="lg:order-last">
          <h2 className="text-3xl md:text-4xl font-bold">
            Make the Bot{" "}
            <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
              Look Like Yours
            </span>
          </h2>
          <p className="text-xl text-muted-foreground mt-4">
            Your bot shows up in Discord with your own name, avatar and status,
            even on the free plan.
          </p>
          <ul className="mt-6 space-y-3">
            {points.map((point) => (
              <li key={point} className="flex gap-3">
                <CheckIcon className="flex-none w-5 h-5 mt-0.5 text-primary" />
                <span>{point}</span>
              </li>
            ))}
          </ul>
          <p className="mt-6 text-muted-foreground">
            See what each plan includes on the{" "}
            <Link href="/pricing" className="text-primary hover:underline">
              pricing page
            </Link>
            .
          </p>
        </div>

        <div className="mx-auto w-full max-w-lg space-y-4">
          <div className="max-w-xs">
            <ProfileMockup statuses={rotatingStatuses} />
          </div>
          <div className="ml-auto max-w-sm">
            <PersonaMockup />
          </div>
        </div>
      </div>
    </section>
  );
}
