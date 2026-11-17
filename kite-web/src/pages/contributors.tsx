import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { withApiData } from "@/lib/apiData";
import env, { isExport } from "@/lib/env/client";
import { formatNumber } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import {
  CodeIcon,
  ExternalLinkIcon,
  GitCommitHorizontalIcon,
  GitPullRequestIcon,
} from "lucide-react";
import { GetStaticProps } from "next";

const REPO = "merlinfuchs/kite";
const CONTRIBUTORS_API = `https://api.github.com/repos/${REPO}/contributors?per_page=100`;
const CREATOR_LOGIN = "merlinfuchs";

export interface Contributor {
  id: number;
  login: string;
  avatar_url: string;
  html_url: string;
  contributions: number;
}

interface GitHubContributor {
  id?: number;
  login?: string;
  avatar_url?: string;
  html_url?: string;
  contributions?: number;
  type?: string;
}

async function fetchContributors(): Promise<Contributor[]> {
  const res = await fetch(CONTRIBUTORS_API, {
    headers: { Accept: "application/vnd.github+json" },
    signal: AbortSignal.timeout(5000),
  });
  if (!res.ok) {
    throw new Error(`GitHub API responded with ${res.status}`);
  }

  const data: GitHubContributor[] = await res.json();
  return data
    .filter(
      (item): item is Contributor & { type: "User" } =>
        item.type === "User" &&
        typeof item.id === "number" &&
        typeof item.login === "string" &&
        typeof item.avatar_url === "string" &&
        typeof item.html_url === "string" &&
        typeof item.contributions === "number"
    )
    .map(({ id, login, avatar_url, html_url, contributions }) => ({
      id,
      login,
      avatar_url,
      html_url,
      contributions,
    }));
}

async function fetchContributorsSafe(): Promise<Contributor[] | null> {
  if (isExport) return null;

  try {
    return await fetchContributors();
  } catch (err) {
    // CI builds and local dev shouldn't fail when GitHub is unreachable; ISR
    // in production rethrows so the previous cached page stays intact.
    if (
      process.env.NODE_ENV !== "production" ||
      process.env.NEXT_PHASE === "phase-production-build"
    ) {
      console.warn(`Failed to fetch GitHub contributors: ${err}`);
      return null;
    }
    throw err;
  }
}

function commitsUrl(login: string) {
  return `https://github.com/${REPO}/commits?author=${encodeURIComponent(
    login
  )}`;
}

function pullsUrl(login: string) {
  return `https://github.com/${REPO}/pulls?q=${encodeURIComponent(
    `is:pr author:${login}`
  )}`;
}

function formatCommits(count: number) {
  return `${formatNumber(count)} ${count === 1 ? "commit" : "commits"}`;
}

export const getStaticProps: GetStaticProps<{
  initial: Contributor[] | null;
}> = async () => {
  const initial = await fetchContributorsSafe();
  return withApiData({ initial });
};

export default function ContributorsPage({
  initial,
}: {
  initial: Contributor[] | null;
}) {
  const query = useQuery({
    queryKey: ["github-contributors"],
    queryFn: fetchContributors,
    enabled: !initial,
  });

  const contributors = initial ?? query.data ?? null;
  const creator = contributors?.find(
    (c) => c.login.toLowerCase() === CREATOR_LOGIN
  );
  const community =
    contributors?.filter((c) => c.login.toLowerCase() !== CREATOR_LOGIN) ?? [];

  return (
    <HomeLayout
      title="Contributors"
      description="The people who build and improve Kite, the free open-source Discord bot maker."
    >
      <section className="border-b bg-muted/30">
        <div className="container max-w-5xl py-14 md:py-20 text-center">
          <h1 className="text-4xl md:text-5xl font-bold">Contributors</h1>
          <p className="text-xl text-muted-foreground mt-4 max-w-2xl mx-auto">
            Kite is open source. Thank you to everyone who has contributed code,
            fixes, and documentation to the project.
          </p>
          <div className="mt-8 flex justify-center">
            <Button variant="outline" asChild>
              <a
                href={env.NEXT_PUBLIC_GITHUB_LINK}
                target="_blank"
                rel="noreferrer noopener"
              >
                <CodeIcon className="w-4 h-4 mr-2" />
                Contribute on GitHub
                <ExternalLinkIcon className="w-3.5 h-3.5 ml-2 text-muted-foreground" />
              </a>
            </Button>
          </div>
        </div>
      </section>

      <div className="container max-w-5xl py-12 space-y-14">
        {!contributors ? (
          query.isError ? (
            <div className="rounded-lg border p-8 text-center">
              <p className="text-muted-foreground">
                Couldn&apos;t load contributors from GitHub right now.
              </p>
              <Button variant="outline" size="sm" className="mt-4" asChild>
                <a
                  href={`https://github.com/${REPO}/graphs/contributors`}
                  target="_blank"
                  rel="noreferrer noopener"
                >
                  View contributors on GitHub
                  <ExternalLinkIcon className="w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                </a>
              </Button>
            </div>
          ) : (
            <div className="space-y-12">
              <section>
                <h2 className="text-2xl font-bold mb-6">Creator</h2>
                <Skeleton className="h-28 w-full rounded-lg" />
              </section>
              <section>
                <h2 className="text-2xl font-bold mb-6">Contributors</h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {Array.from({ length: 6 }).map((_, i) => (
                    <Skeleton key={i} className="h-20 w-full rounded-lg" />
                  ))}
                </div>
              </section>
            </div>
          )
        ) : (
          <>
            {creator && (
              <section>
                <h2 className="text-2xl font-bold mb-6">Creator</h2>
                <div className="rounded-lg border bg-muted/40 p-6 flex flex-col sm:flex-row sm:items-center gap-5">
                  <a
                    href={creator.html_url}
                    target="_blank"
                    rel="noreferrer noopener"
                    className="flex-none self-start sm:self-center"
                  >
                    <Avatar className="h-16 w-16 border">
                      <AvatarImage
                        src={creator.avatar_url}
                        alt={creator.login}
                      />
                      <AvatarFallback>MF</AvatarFallback>
                    </Avatar>
                  </a>

                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2.5">
                      <a
                        href={creator.html_url}
                        target="_blank"
                        rel="noreferrer noopener"
                        className="text-xl font-bold hover:text-primary transition-colors"
                      >
                        Merlin Fuchs
                      </a>
                      <Badge className="bg-primary/15 text-primary hover:bg-primary/15 border-0">
                        Creator &amp; maintainer
                      </Badge>
                    </div>
                    <div className="text-sm text-muted-foreground mt-1 flex flex-wrap items-center gap-x-2">
                      <span>@{creator.login}</span>
                      <span>&middot;</span>
                      <a
                        href={commitsUrl(creator.login)}
                        target="_blank"
                        rel="noreferrer noopener"
                        className="inline-flex items-center gap-1 hover:text-foreground transition-colors"
                        data-indexnow-ignore
                      >
                        <GitCommitHorizontalIcon className="w-3.5 h-3.5" />
                        {formatCommits(creator.contributions)}
                      </a>
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center gap-2 flex-none">
                    <Button variant="outline" size="sm" asChild>
                      <a
                        href={commitsUrl(creator.login)}
                        target="_blank"
                        rel="noreferrer noopener"
                      >
                        Commits
                        <ExternalLinkIcon className="w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                      </a>
                    </Button>
                    <Button variant="outline" size="sm" asChild>
                      <a
                        href={pullsUrl(creator.login)}
                        target="_blank"
                        rel="noreferrer noopener"
                      >
                        PRs
                        <ExternalLinkIcon className="w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                      </a>
                    </Button>
                    <Button variant="outline" size="sm" asChild>
                      <a
                        href={creator.html_url}
                        target="_blank"
                        rel="noreferrer noopener"
                      >
                        GitHub
                        <ExternalLinkIcon className="w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                      </a>
                    </Button>
                  </div>
                </div>
              </section>
            )}

            {community.length > 0 && (
              <section>
                <div className="flex items-baseline justify-between gap-4 mb-6">
                  <h2 className="text-2xl font-bold">Contributors</h2>
                  <span
                    className="text-sm text-muted-foreground"
                    data-indexnow-ignore
                  >
                    {community.length}{" "}
                    {community.length === 1 ? "contributor" : "contributors"}
                  </span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {community.map((contributor) => (
                    <div
                      key={contributor.id}
                      className="group relative flex items-center gap-3.5 rounded-lg border bg-card p-4 transition-colors hover:border-primary"
                    >
                      <Avatar className="h-12 w-12 flex-none border">
                        <AvatarImage
                          src={contributor.avatar_url}
                          alt={contributor.login}
                        />
                        <AvatarFallback>
                          {contributor.login.slice(0, 2).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>

                      <div className="min-w-0 flex-1">
                        <a
                          href={contributor.html_url}
                          target="_blank"
                          rel="noreferrer noopener"
                          className="font-semibold block truncate group-hover:text-primary transition-colors after:absolute after:inset-0"
                        >
                          @{contributor.login}
                        </a>
                        <div className="relative z-10 flex flex-wrap items-center gap-x-2 text-sm text-muted-foreground">
                          <a
                            href={commitsUrl(contributor.login)}
                            target="_blank"
                            rel="noreferrer noopener"
                            className="inline-flex items-center gap-1 hover:text-foreground transition-colors"
                            data-indexnow-ignore
                          >
                            <GitCommitHorizontalIcon className="w-3.5 h-3.5" />
                            {formatCommits(contributor.contributions)}
                          </a>
                          <span aria-hidden="true">&middot;</span>
                          <a
                            href={pullsUrl(contributor.login)}
                            target="_blank"
                            rel="noreferrer noopener"
                            className="inline-flex items-center gap-1 hover:text-foreground transition-colors"
                          >
                            <GitPullRequestIcon className="w-3.5 h-3.5" />
                            PRs
                          </a>
                        </div>
                      </div>

                      <ExternalLinkIcon className="w-4 h-4 flex-none text-muted-foreground opacity-60 group-hover:opacity-100 group-hover:text-primary transition-all" />
                    </div>
                  ))}
                </div>
              </section>
            )}
          </>
        )}
      </div>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
