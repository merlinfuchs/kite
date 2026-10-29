import { GetServerSideProps } from "next";

export default function AppContextMenusRedirect() {
  return null;
}

export const getServerSideProps: GetServerSideProps = async ({ params }) => {
  const appId = params?.appId;
  if (typeof appId !== "string") {
    return { notFound: true };
  }

  return {
    redirect: {
      destination: `/apps/${encodeURIComponent(appId)}/commands`,
      permanent: false,
    },
  };
};
