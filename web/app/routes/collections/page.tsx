import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { Route } from "./+types/page";
import { columnsQuery, rowsQuery, prepareColumns } from "./columns";
import { useState, useMemo, useCallback } from "react";
import { RefreshButton } from "~/components/shared/refresh-button";
import { SearchDataTableWrapper } from "~/components/shared/search-data-table-wrapper";
import { DataTableSkeleton } from "~/components/shared/data-table-skeleton";
import { DeleteButton } from "~/components/shared/delete-button";
import { useNavigate, useOutletContext } from "react-router";
import type { ProjectLayoutOutletContext } from "~/components/shared/project-layout";

const DEFAULT_PAGE_SIZE = 50;
const DEFAULT_PAGE_INDEX = 0;

type PaginationType = {
  pageIndex: number;
  pageSize: number;
};

export default function CollectionPageContent({
  params,
}: Route.ComponentProps) {
  const { projectId, collectionId } = params;
  const { projectDetails, services } =
    useOutletContext<ProjectLayoutOutletContext>();

  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [pagination, setPagination] = useState<PaginationType>({
    pageIndex: DEFAULT_PAGE_INDEX,
    pageSize: DEFAULT_PAGE_SIZE,
  });

  const {
    isLoading: isColumnsLoading,
    data: columnsData = [],
    isFetching: isColumnsFetching,
    error: columnsError,
  } = useQuery(columnsQuery(projectId, collectionId)) || { data: [] };

  const columns = useMemo(() => {
    if (!columnsData || !Array.isArray(columnsData)) {
      return [];
    }

    return prepareColumns(columnsData, collectionId);
  }, [columnsData, collectionId]);

  const [filterParams, setFilterParams] = useState<Record<string, string>>({});

  // Handle filter changes with pagination reset
  const handleFilterChange = useCallback(
    (params: Record<string, string>) => {
      setFilterParams(params);
      // Reset to first page when filters change
      if (pagination.pageIndex !== 0) {
        setPagination({
          ...pagination,
          pageIndex: 0,
        });
      }
    },
    [pagination, setPagination]
  );

  const resetFilters = useCallback(() => {
    setFilterParams({});
  }, []);

  const {
    isLoading: isRowsLoading,
    data: rowsData = { totalCount: 0, rows: [] },
    isFetching: isRowsFetching,
    error: rowsError,
  } = useQuery({
    ...rowsQuery(
      projectId,
      projectDetails?.dbName as string,
      collectionId,
      pagination,
      filterParams
    ),
  });

  // Safely destructure to handle undefined
  const { rows = [], totalCount = 0 } = rowsData;

  // Track initial loading vs pagination loading separately
  const isInitialLoading = isColumnsLoading || isRowsLoading;
  const isFetching = isColumnsFetching || isRowsFetching;

  const handleRefresh = async () => {
    if (collectionId) {
      // Invalidate queries
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["columns", projectId, collectionId],
        }),
        queryClient.invalidateQueries({
          queryKey: [
            "rows",
            projectId,
            collectionId,
            pagination.pageSize,
            pagination.pageIndex,
            filterParams,
          ],
        }),
      ]);
    }
  };

  const onPaginationChange = useCallback(
    (
      updaterOrValue: PaginationType | ((old: PaginationType) => PaginationType)
    ) => {
      const newPagination =
        typeof updaterOrValue === "function"
          ? updaterOrValue(pagination)
          : updaterOrValue;
      setPagination(newPagination);
      queryClient.invalidateQueries({
        queryKey: [
          "rows",
          projectId,
          collectionId,
          newPagination.pageSize,
          newPagination.pageIndex,
          filterParams,
        ],
      });
    },
    [pagination, projectId, collectionId, filterParams, queryClient]
  );

  const handleDeleteTable = useCallback(async () => {
    if (!collectionId || !projectId) return;

    const confirmDelete = window.confirm(
      `Are you sure you want to delete the table "${collectionId}"? This action cannot be undone.`
    );

    if (!confirmDelete) return;

    const response = await services.collections.deleteCollection(
      projectId,
      collectionId
    );

    if (response.ok) {
      // Invalidate collections query to refresh the sidebar
      await queryClient.invalidateQueries({
        queryKey: ["collections", projectId],
      });

      // Navigate back to collections without specific collection
      navigate(`/projects/${projectId}/collections`);
    } else {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || "Failed to delete table");
    }
  }, [collectionId, projectId, queryClient, navigate]);

  const noCollectionSelected = !collectionId;

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="border-b px-4 py-2 mb-2 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="text-base font-bold text-foreground h-[32px] flex flex-col justify-center">
            Collections / {collectionId && `${collectionId}`}
          </div>
          <div className="flex items-center gap-2">
            {collectionId && (
              <DeleteButton onDelete={handleDeleteTable} title="Delete Table" />
            )}
            <RefreshButton
              onRefresh={useCallback(handleRefresh, [
                collectionId,
                projectId,
                queryClient,
              ])}
              title="Refresh Collections and Collection Data"
            />
          </div>
        </div>
      </div>

      {isInitialLoading && (
        <div className="rounded-md border mx-4 py-4">
          <DataTableSkeleton columns={5} rows={8} />
        </div>
      )}

      {!isInitialLoading &&
        Array.isArray(columns) &&
        columns.length > 0 &&
        collectionId && (
          <div className="flex-1 min-h-0 py-2 pb-3 overflow-hidden">
            <SearchDataTableWrapper
              columns={columns}
              rawColumns={columnsData}
              data={Array.isArray(rows) ? rows : []}
              isFetching={isFetching}
              emptyMessage="No table data found."
              className="w-full h-full"
              pagination={pagination}
              totalRows={totalCount}
              projectId={projectId}
              collectionId={collectionId}
              onFilterChange={handleFilterChange}
              onPaginationChange={onPaginationChange}
            />
          </div>
        )}

      {!isInitialLoading &&
        (!Array.isArray(columns) || columns.length === 0) &&
        collectionId && (
          <div className="flex-1 min-h-0 flex items-center justify-center mx-4">
            <div className="text-md text-muted-foreground border rounded-md p-8 bg-muted/10">
              No Table Data Found
            </div>
          </div>
        )}

      {noCollectionSelected && !isInitialLoading && (
        <div className="flex-1 min-h-0 flex items-center justify-center mx-4">
          <div className="text-md text-muted-foreground border rounded-md p-8 bg-muted/10">
            Please select a collection from the sidebar
          </div>
        </div>
      )}
    </div>
  );
}
