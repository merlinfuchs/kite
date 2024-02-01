import { QueryCache, QueryClient } from "react-query";
//import { useToasts } from "../util/toasts";

const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (err) => {
      /*useToasts.getState().create({
        type: "error",
        title: "Unexpect API error",
        message: `${err}`,
      });*/
    },
  }),
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: (failureCount, err: any) => {
        if (failureCount >= 3) {
          return false;
        }
        return err.status >= 500;
      },
      staleTime: 1000 * 60 * 3,
    },
  },
});

export default queryClient;
