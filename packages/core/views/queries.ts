import { queryOptions } from "@tanstack/react-query";
import { api } from "../api";
import type { ListViewsParams } from "../types";

export const viewKeys = {
  all: (wsId: string) => ["views", wsId] as const,
  list: (wsId: string, params: ListViewsParams) =>
    [...viewKeys.all(wsId), "list", params] as const,
};

export function viewListOptions(wsId: string, params: ListViewsParams) {
  return queryOptions({
    queryKey: viewKeys.list(wsId, params),
    queryFn: () => api.listViews(params),
    select: (data) => data.views,
  });
}
