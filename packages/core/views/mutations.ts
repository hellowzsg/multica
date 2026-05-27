import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api";
import { viewKeys } from "./queries";
import { useWorkspaceId } from "../hooks";
import type {
  CreateViewRequest,
  UpdateViewRequest,
  ReorderViewsRequest,
  ListViewsResponse,
  ListViewsParams,
} from "../types";

export function useCreateView(params: ListViewsParams) {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (data: CreateViewRequest) => api.createView(data),
    onSuccess: (view) => {
      qc.setQueryData<ListViewsResponse>(viewKeys.list(wsId, params), (old) =>
        old && !old.views.some((v) => v.id === view.id)
          ? { ...old, views: [...old.views, view], total: old.total + 1 }
          : old,
      );
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: viewKeys.all(wsId) });
    },
  });
}

export function useUpdateView(params: ListViewsParams) {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & UpdateViewRequest) =>
      api.updateView(id, data),
    onMutate: async ({ id, ...data }) => {
      await qc.cancelQueries({ queryKey: viewKeys.list(wsId, params) });
      const prev = qc.getQueryData<ListViewsResponse>(viewKeys.list(wsId, params));
      qc.setQueryData<ListViewsResponse>(viewKeys.list(wsId, params), (old) =>
        old
          ? {
              ...old,
              views: old.views.map((v) =>
                v.id === id ? { ...v, ...data } : v,
              ),
            }
          : old,
      );
      return { prev };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) qc.setQueryData(viewKeys.list(wsId, params), ctx.prev);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: viewKeys.all(wsId) });
    },
  });
}

export function useDeleteView(params: ListViewsParams) {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (id: string) => api.deleteView(id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: viewKeys.list(wsId, params) });
      const prev = qc.getQueryData<ListViewsResponse>(viewKeys.list(wsId, params));
      qc.setQueryData<ListViewsResponse>(viewKeys.list(wsId, params), (old) =>
        old
          ? { ...old, views: old.views.filter((v) => v.id !== id), total: old.total - 1 }
          : old,
      );
      return { prev };
    },
    onError: (_err, _id, ctx) => {
      if (ctx?.prev) qc.setQueryData(viewKeys.list(wsId, params), ctx.prev);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: viewKeys.all(wsId) });
    },
  });
}

export function useReorderViews(params: ListViewsParams) {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (data: ReorderViewsRequest) => api.reorderViews(data),
    onMutate: async (data) => {
      await qc.cancelQueries({ queryKey: viewKeys.list(wsId, params) });
      const prev = qc.getQueryData<ListViewsResponse>(viewKeys.list(wsId, params));
      qc.setQueryData<ListViewsResponse>(viewKeys.list(wsId, params), (old) => {
        if (!old) return old;
        const posMap = new Map(data.items.map((i) => [i.id, i.position]));
        const views = old.views
          .map((v) => (posMap.has(v.id) ? { ...v, position: posMap.get(v.id)! } : v))
          .sort((a, b) => a.position - b.position);
        return { ...old, views };
      });
      return { prev };
    },
    onError: (_err, _data, ctx) => {
      if (ctx?.prev) qc.setQueryData(viewKeys.list(wsId, params), ctx.prev);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: viewKeys.all(wsId) });
    },
  });
}
