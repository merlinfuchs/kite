import logo from "@/assets/logo/orange@1024.png";
import Link from "next/link";
import env from "@/lib/env/client";

export default function HomeFooter() {
  return (
    <footer id="footer">
      <hr className="w-11/12 mx-auto" />

      <section className="container py-20 grid grid-cols-2 md:grid-cols-4 xl:grid-cols-6 gap-x-12 gap-y-8">
        <div className="col-span-full xl:col-span-2">
          <a
            rel="noreferrer noopener"
            href="/"
            className="font-bold text-xl flex items-center"
          >
            <img src={logo.src} className="h-10 w-10 mr-3" />
            <div>Kite</div>
          </a>
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="font-bold text-lg">Get in touch</h3>
          <div>
            <a
              rel="noreferrer noopener"
              href={env.NEXT_PUBLIC_GITHUB_LINK}
              target="_blank"
              className="opacity-60 hover:opacity-100"
            >
              Github
            </a>
          </div>

          <div>
            <a
              rel="noreferrer noopener"
              href={env.NEXT_PUBLIC_DISCORD_LINK}
              target="_blank"
              className="opacity-60 hover:opacity-100"
            >
              Discord
            </a>
          </div>

          <div>
            <a
              rel="noreferrer noopener"
              href={`mailto:${env.NEXT_PUBLIC_CONTACT_EMAIL}`}
              target="_blank"
              className="opacity-60 hover:opacity-100"
            >
              Email
            </a>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="font-bold text-lg">Resources</h3>
          <div>
            <a href="/docs" className="opacity-60 hover:opacity-100">
              Documentation
            </a>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="font-bold text-lg">Legal</h3>
          <div>
            <Link href="/terms" className="opacity-60 hover:opacity-100">
              Terms of Service
            </Link>
          </div>

          <div>
            <Link href="/privacy" className="opacity-60 hover:opacity-100">
              Privacy Policy
            </Link>
          </div>
        </div>
      </section>

      <section className="container pb-14 text-center">
        <h3>
          Copyright 2024 &copy; made by{" "}
          <a
            target="_blank"
            href="https://merlin.gg"
            className="text-primary transition-all border-primary hover:border-b-2"
          >
            Merlin Fuchs
          </a>
        </h3>
      </section>
    </footer>
  );
}
