import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "./client";
import {
  MarketplaceListingCreateRequest,
  MarketplaceListingCreateResponse,
  MarketplaceListingDeleteResponse,
  MarketplaceListingGetResponse,
  MarketplaceListingImportResponse,
  MarketplaceListingListResponse,
  MarketplaceListingReviewRequest,
  MarketplaceListingReviewResponse,
  MarketplaceListingUpdateRequest,
  MarketplaceListingUpdateResponse,
  MarketplaceMeGetResponse,
  MarketplaceModeratorCreateRequest,
  MarketplaceModeratorCreateResponse,
  MarketplaceModeratorDeleteResponse,
  MarketplaceModeratorListResponse,
  MarketplaceReportCreateRequest,
  MarketplaceReportCreateResponse,
  MarketplaceReportListResponse,
  MarketplaceReportResolveResponse,
} from "../types/wire.gen";

export type MarketplaceListingKind =
  | ""
  | "command"
  | "event_listener"
  | "message"
  | "module";
export type MarketplaceListingSort = "popular" | "recent";
export type MarketplaceListingStatus =
  | "pending"
  | "approved"
  | "rejected"
  | "removed";

export interface MarketplaceListingsParams {
  search?: string;
  kind?: MarketplaceListingKind;
  sort?: MarketplaceListingSort;
  limit?: number;
  offset?: number;
}

function listingsQueryString(
  params: MarketplaceListingsParams & { status?: string }
) {
  const query = new URLSearchParams();
  if (params.search) query.set("search", params.search);
  if (params.kind) query.set("kind", params.kind);
  if (params.sort) query.set("sort", params.sort);
  if (params.status) query.set("status", params.status);
  if (params.limit) query.set("limit", params.limit.toString());
  if (params.offset) query.set("offset", params.offset.toString());
  const res = query.toString();
  return res ? `?${res}` : "";
}

export function useMarketplaceMeQuery() {
  return useQuery({
    queryKey: ["marketplace", "me"],
    queryFn: () => apiRequest<MarketplaceMeGetResponse>(`/v1/marketplace/me`),
  });
}

export function useMarketplaceListingsQuery(params: MarketplaceListingsParams) {
  return useQuery({
    queryKey: ["marketplace", "listings", "browse", params],
    queryFn: () =>
      apiRequest<MarketplaceListingListResponse>(
        `/v1/marketplace/listings${listingsQueryString(params)}`
      ),
    staleTime: 1000 * 60,
  });
}

export function useMarketplaceMyListingsQuery() {
  return useQuery({
    queryKey: ["marketplace", "listings", "mine"],
    queryFn: () =>
      apiRequest<MarketplaceListingListResponse>(
        `/v1/marketplace/listings/@me`
      ),
  });
}

export function useMarketplaceListingQuery(listingId: string | null) {
  return useQuery({
    queryKey: ["marketplace", "listings", "get", listingId],
    queryFn: () =>
      apiRequest<MarketplaceListingGetResponse>(
        `/v1/marketplace/listings/${listingId}`
      ),
    enabled: !!listingId,
  });
}

export function useMarketplaceModerationListingsQuery(
  status: MarketplaceListingStatus,
  enabled = true
) {
  return useQuery({
    queryKey: ["marketplace", "listings", "moderation", status],
    queryFn: () =>
      apiRequest<MarketplaceListingListResponse>(
        `/v1/marketplace/moderation/listings${listingsQueryString({
          status,
          sort: "recent",
          limit: 50,
        })}`
      ),
    enabled,
  });
}

export function useMarketplaceReportsQuery(enabled = true) {
  return useQuery({
    queryKey: ["marketplace", "reports"],
    queryFn: () =>
      apiRequest<MarketplaceReportListResponse>(
        `/v1/marketplace/moderation/reports`
      ),
    enabled,
  });
}

export function useMarketplaceModeratorsQuery(enabled = true) {
  return useQuery({
    queryKey: ["marketplace", "moderators"],
    queryFn: () =>
      apiRequest<MarketplaceModeratorListResponse>(
        `/v1/marketplace/moderation/moderators`
      ),
    enabled,
  });
}

function useInvalidateListings() {
  const client = useQueryClient();
  return () => {
    client.invalidateQueries({ queryKey: ["marketplace", "listings"] });
    client.invalidateQueries({ queryKey: ["marketplace", "reports"] });
  };
}

export function useMarketplaceListingCreateMutation() {
  const invalidate = useInvalidateListings();

  return useMutation({
    mutationFn: (req: MarketplaceListingCreateRequest) =>
      apiRequest<MarketplaceListingCreateResponse>(`/v1/marketplace/listings`, {
        method: "POST",
        body: JSON.stringify(req),
        headers: {
          "Content-Type": "application/json",
        },
      }),
    onSuccess: invalidate,
  });
}

export function useMarketplaceListingUpdateMutation(listingId: string) {
  const invalidate = useInvalidateListings();

  return useMutation({
    mutationFn: (req: MarketplaceListingUpdateRequest) =>
      apiRequest<MarketplaceListingUpdateResponse>(
        `/v1/marketplace/listings/${listingId}`,
        {
          method: "PATCH",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: invalidate,
  });
}

export function useMarketplaceListingDeleteMutation() {
  const invalidate = useInvalidateListings();

  return useMutation({
    mutationFn: (listingId: string) =>
      apiRequest<MarketplaceListingDeleteResponse>(
        `/v1/marketplace/listings/${listingId}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: invalidate,
  });
}

// Returns the listing with its flows, the items are then imported with the
// regular command and event listener import mutations.
export function useMarketplaceListingImportMutation() {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (listingId: string) =>
      apiRequest<MarketplaceListingImportResponse>(
        `/v1/marketplace/listings/${listingId}/import`,
        {
          method: "POST",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({
        queryKey: ["marketplace", "listings", "browse"],
      });
    },
  });
}

export function useMarketplaceReportCreateMutation(listingId: string) {
  return useMutation({
    mutationFn: (req: MarketplaceReportCreateRequest) =>
      apiRequest<MarketplaceReportCreateResponse>(
        `/v1/marketplace/listings/${listingId}/reports`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
  });
}

export function useMarketplaceListingReviewMutation(listingId: string) {
  const invalidate = useInvalidateListings();

  return useMutation({
    mutationFn: (req: MarketplaceListingReviewRequest) =>
      apiRequest<MarketplaceListingReviewResponse>(
        `/v1/marketplace/moderation/listings/${listingId}/review`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: invalidate,
  });
}

export function useMarketplaceReportResolveMutation() {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (reportId: string) =>
      apiRequest<MarketplaceReportResolveResponse>(
        `/v1/marketplace/moderation/reports/${reportId}/resolve`,
        {
          method: "POST",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["marketplace", "reports"] });
    },
  });
}

export function useMarketplaceModeratorCreateMutation() {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (req: MarketplaceModeratorCreateRequest) =>
      apiRequest<MarketplaceModeratorCreateResponse>(
        `/v1/marketplace/moderation/moderators`,
        {
          method: "POST",
          body: JSON.stringify(req),
          headers: {
            "Content-Type": "application/json",
          },
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["marketplace", "moderators"] });
    },
  });
}

export function useMarketplaceModeratorDeleteMutation() {
  const client = useQueryClient();

  return useMutation({
    mutationFn: (discordUserId: string) =>
      apiRequest<MarketplaceModeratorDeleteResponse>(
        `/v1/marketplace/moderation/moderators/${discordUserId}`,
        {
          method: "DELETE",
        }
      ),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["marketplace", "moderators"] });
    },
  });
}
