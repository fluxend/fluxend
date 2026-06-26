import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { Route } from "./+types/page";
import { columnsQuery, rowsQuery, prepareColumns } from "./columns";
import { useState, useMemo, useCallback, useRef, useEffect } from "react";
import { RefreshButton } from "~/components/shared/refresh-button";
import { SearchDataTableWrapper } from "~/components/shared/search-data-table-wrapper";
import { DataTableSkeleton } from "~/components/shared/data-table-skeleton";
import { useNavigate, useOutletContext } from "react-router";
import type { ProjectLayoutOutletContext } from "~/components/shared/project-layout";
import type { RowSelectionState } from "@tanstack/react-table";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "~/components/ui/alert-dialog";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Tables - Fluxend" },
    { name: "description", content: "Manage your database tables" },
  ];
}
import { Button } from "~/components/ui/button";
import { Trash2, Pencil, FileText, Webhook } from "lucide-react";
import { toast } from "sonner";
import type { TableRow, CellValue } from "~/types/table";

const DEFAULT_PAGE_SIZE = 50;
const DEFAULT_PAGE_INDEX = 0;

type PaginationType = {
  pageIndex: number;
  pageSize: number;
};

export default function TablePageContent({ params }: Route.ComponentProps) {
  const { projectId, tableId } = params;
  const { projectDetails, services } =
    useOutletContext<ProjectLayoutOutletContext>();

  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [pagination, setPagination] = useState<PaginationType>({
    pageIndex: DEFAULT_PAGE_INDEX,
    pageSize: DEFAULT_PAGE_SIZE,
  });
  const [selectedRows, setSelectedRows] = useState<TableRow[]>([]);
  const [rowSelectionState, setRowSelectionState] = useState<RowSelectionState>({});
  const [isDeletingRows, setIsDeletingRows] = useState(false);
  const [showBulkDeleteDialog, setShowBulkDeleteDialog] = useState(false);

  const {
    isLoading: isColumnsLoading,
    data: columnsData = [],
    isFetching: isColumnsFetching,
    error: columnsError,
  } = useQuery(columnsQuery(projectId, tableId)) || { data: [] };

  const handleCellUpdate = useCallback(async (
    rowId: string,
    columnName: string,
    value: CellValue,
    rowData: TableRow
  ): Promise<boolean> => {
    try {
      const updateData = {
        [columnName]: value
      };

      const response = await services.tables.updateTableRow(
        projectId,
        tableId,
        rowId,
        updateData,
      );

      if (response.ok) {
        // Don't show toast or invalidate queries - just return success
        // The optimistic update will remain in place
        return true;
      } else {
        // Try to extract error message from response
        try {
          const errorData = await response.json();
          const errorMessage = errorData?.message || errorData?.error || errorData?.detail || 'Failed to update';
          toast.error(errorMessage);
        } catch {
          toast.error('Failed to update');
        }
        // Return false to trigger revert in OptimisticTableCell
        return false;
      }
    } catch (error) {
      // Handle network errors or other exceptions
      const errorMessage = error instanceof Error ? error.message : 'Network error occurred';
      toast.error(errorMessage);
      // Return false to trigger revert in OptimisticTableCell
      return false;
    }
  }, [projectId, tableId, services.tables]);

  const columns = useMemo(() => {
    if (!columnsData || !Array.isArray(columnsData)) {
      return [];
    }

    return prepareColumns(columnsData, tableId);
  }, [columnsData, tableId]);

  const [filterParams, setFilterParams] = useState<Record<string, string>>({});
  const [searchQuery, setSearchQuery] = useState<string>("");

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
      tableId,
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
    if (tableId) {
      // Invalidate queries
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["columns", projectId, tableId],
        }),
        queryClient.invalidateQueries({
          queryKey: [
            "rows",
            projectId,
            tableId,
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
          tableId,
          newPagination.pageSize,
          newPagination.pageIndex,
          filterParams,
        ],
      });
    },
    [pagination, projectId, tableId, filterParams, queryClient]
  );

  const handleDeleteTable = useCallback(async () => {
    if (!tableId || !projectId) return;

    const response = await services.tables.deleteTable(projectId, tableId);

    if (response.ok) {
      // Invalidate collections query to refresh the sidebar
      await queryClient.invalidateQueries({
        queryKey: ["tables", projectId],
      });

      navigate(`/projects/${projectId}/tables`);
      toast.success(`Table ${tableId} deleted successfully`);
    } else if (response?.errors) {
      toast.error(response?.errors[0]);
    } else {
      // Show generic error message
      toast.error('Failed to delete table');
    }
  }, [tableId, projectId, queryClient, navigate]);

  const handleBulkDelete = useCallback(async () => {
    if (!tableId || !projectId || selectedRows.length === 0) return;

    const rowIds = selectedRows.map(row => row.id).filter(Boolean);
    if (rowIds.length === 0) {
      toast.error('No valid rows selected for deletion');
      return;
    }

    setIsDeletingRows(true);

    try {
      const response = await services.tables.deleteTableRows(
        projectId,
        tableId,
        rowIds,
      );

      if (response.ok) {
        // Invalidate rows query to refresh the table data
        await queryClient.invalidateQueries({
          queryKey: [
            "rows",
            projectId,
            tableId,
            pagination.pageSize,
            pagination.pageIndex,
            filterParams,
          ],
        });

        toast.success(`Successfully deleted ${rowIds.length} row${rowIds.length > 1 ? 's' : ''}`);
        setSelectedRows([]); // Clear selection after successful deletion
        setRowSelectionState({}); // Clear the selection state
        setShowBulkDeleteDialog(false);
      } else {
        // Try to extract error message from response
        try {
          const errorData = await response.json();
          const errorMessage = errorData?.message || errorData?.error || errorData?.detail || 'Failed to delete rows';
          toast.error(errorMessage);
        } catch {
          toast.error('Failed to delete rows');
        }
      }
    } catch (error) {
      // Handle network errors or other exceptions
      const errorMessage = error instanceof Error ? error.message : 'Network error occurred';
      toast.error(errorMessage);
    } finally {
      setIsDeletingRows(false);
    }
  }, [tableId, projectId, queryClient, pagination, filterParams, services.tables, selectedRows]);

  const noTableSelected = !tableId;

  const tableMeta = useMemo(() => ({
    onCellUpdate: handleCellUpdate,
  }), [handleCellUpdate]);

  // Clear selection when tableId changes
  useEffect(() => {
    setSelectedRows([]);
    setRowSelectionState({});
  }, [tableId]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd/Ctrl + A to select all
      if ((e.metaKey || e.ctrlKey) && e.key === 'a' && !e.shiftKey) {
        const allRowsSelection: RowSelectionState = {};
        rows.forEach((_, index) => {
          allRowsSelection[index] = true;
        });
        setRowSelectionState(allRowsSelection);
        e.preventDefault();
      }
      
      // Delete key to open delete dialog when rows are selected
      if (e.key === 'Delete' && selectedRows.length > 0 && !isDeletingRows) {
        setShowBulkDeleteDialog(true);
        e.preventDefault();
      }
      
      // Escape to clear selection
      if (e.key === 'Escape' && selectedRows.length > 0) {
        setRowSelectionState({});
        setSelectedRows([]);
        e.preventDefault();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [selectedRows.length, rows, isDeletingRows]);

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="border-b px-4 py-2 mb-2 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="text-base font-bold text-foreground h-[32px] flex flex-col justify-center">
              Tables / {tableId && `${tableId}`}
            </div>
            {selectedRows.length > 0 && (
              <div className="text-sm text-muted-foreground">
                ({selectedRows.length} selected)
              </div>
            )}
          </div>
          <div className="flex items-center gap-2">
            {!noTableSelected && (
              <>
                {selectedRows.length > 0 && (
                  <>
                    <Button
                      type="button"
                      variant="destructive"
                      size="sm"
                      className="cursor-pointer"
                      onClick={() => setShowBulkDeleteDialog(true)}
                      disabled={isDeletingRows}
                      title={`Delete ${selectedRows.length} selected row${selectedRows.length > 1 ? 's' : ''}`}
                      aria-label={`Delete ${selectedRows.length} selected row${selectedRows.length > 1 ? 's' : ''}`}
                    >
                      <Trash2 className="h-4 w-4 mr-2" aria-hidden="true" />
                      Delete {selectedRows.length} row{selectedRows.length > 1 ? 's' : ''}
                    </Button>
                    <AlertDialog open={showBulkDeleteDialog} onOpenChange={setShowBulkDeleteDialog}>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>
                            Delete {selectedRows.length} row{selectedRows.length > 1 ? 's' : ''}?
                          </AlertDialogTitle>
                          <AlertDialogDescription>
                            This action cannot be undone. This will permanently delete
                            <strong className="text-destructive"> {selectedRows.length} </strong>
                            selected row{selectedRows.length > 1 ? 's' : ''} from the table.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel disabled={isDeletingRows}>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            onClick={handleBulkDelete}
                            className="cursor-pointer bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            disabled={isDeletingRows}
                          >
                            {isDeletingRows ? "Deleting..." : "Delete"}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </>
                )}
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="cursor-pointer"
                  onClick={() => navigate(`/projects/${projectId}/docs?table=${tableId}`)}
                  title="View API Docs"
                >
                  <FileText />
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="cursor-pointer"
                  onClick={() => navigate(`/projects/${projectId}/tables/${tableId}/webhooks`)}
                  title="Webhooks"
                >
                  <Webhook />
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="cursor-pointer"
                  onClick={() => navigate(`/projects/${projectId}/tables/${tableId}/edit`)}
                  title="Edit Table"
                >
                  <Pencil />
                </Button>
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="cursor-pointer"
                    >
                      <Trash2 />
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>
                        Are you absolutely sure?
                      </AlertDialogTitle>
                      <AlertDialogDescription>
                        This action cannot be undone. This will permanently delete
                        table{" "}
                        <strong className="text-destructive">{tableId}</strong>{" "}
                        from our servers.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={handleDeleteTable}
                        className="cursor-pointer"
                      >
                        Continue
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </>
            )}
            <RefreshButton
              onRefresh={useCallback(handleRefresh, [
                tableId,
                projectId,
                queryClient,
              ])}
              size="sm"
              title="Refresh Tables and Table Data"
            />
          </div>
        </div>
      </div>

      {isInitialLoading && (
        <div className="rounded-lg border mx-4 py-4">
          <DataTableSkeleton columns={5} rows={8} />
        </div>
      )}

      {!isInitialLoading &&
        Array.isArray(columns) &&
        columns.length > 0 &&
        tableId && (
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
              tableId={tableId}
              onFilterChange={handleFilterChange}
              searchQuery={searchQuery}
              onSearchQueryChange={setSearchQuery}
              onPaginationChange={onPaginationChange}
              tableMeta={tableMeta}
              onRowSelectionChange={setSelectedRows}
              rowSelection={rowSelectionState}
              onRowSelectionStateChange={setRowSelectionState}
            />
          </div>
        )}

      {!isInitialLoading &&
        (!Array.isArray(columns) || columns.length === 0) &&
        tableId && (
          <div className="flex-1 min-h-0 flex items-center justify-center mx-4">
            <div className="text-md text-muted-foreground border rounded-lg p-8 bg-muted/10">
              No Table Data Found
            </div>
          </div>
        )}

      {noTableSelected && !isInitialLoading && (
        <div className="flex-1 min-h-0 flex items-center justify-center mx-4">
          <div className="text-md text-muted-foreground border rounded-lg p-8 bg-muted/10">
            Please select a collection from the sidebar
          </div>
        </div>
      )}

      {/* EditRowSheet is no longer needed as we're using inline editing */}
    </div>
  );
}
