import { useQuery } from "react-query";
import { GuildListResponse } from "./wire";

export function useGuildsQuery() {
  return useQuery<GuildListResponse>(["guilds"], () => {
    return fetch(`/api/v1/guilds`).then((res) => res.json());
  });
}
