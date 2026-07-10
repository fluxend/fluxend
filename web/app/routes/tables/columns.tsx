import {
  Hash,
  AlignLeft,
  Clock,
  Text,
  MoreVertical,
  RotateCcw,
  Circle,
  ToggleLeft,
  Calendar,
  FileJson,
  Shuffle,
  Save,
  X,
  Edit2,
  Eye,
  EyeOff,
} from "lucide-react";
import type { ColumnDef, PaginationState, Row } from "@tanstack/react-table";
import { Button } from "~/components/ui/button";
import { Checkbox } from "~/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from "~/components/ui/dropdown-menu";

import React, { useState } from "react";
import { Switch } from "~/components/ui/switch";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "~/components/ui/tooltip";
import { formatTimestamp, getProjectApiUrl, getTypedResponseData } from "~/lib/utils";
import type { APIResponse } from "~/lib/types";
import { getClientAuthToken } from "~/lib/auth";
import { initializeServices } from "~/services";
import { OptimisticTableCell } from "~/components/tables/optimistic-table-cell";
import type { CellValue } from "~/types/table";

export const columnsQuery = (projectId: string, tableId: string) => ({
  queryKey: ["columns", projectId, tableId],
  queryFn: async () => {
    const authToken = await getClientAuthToken();
    if (!authToken) {
      throw new Error("Unauthorized");
    }

    const services = initializeServices(authToken);

    const res = await services.tables.getTableColumns(projectId, tableId);

    if (!res.ok && res.status > 500) {
      throw new Error("Unexpected Error");
    }

    const data = await getTypedResponseData<APIResponse<any>>(res);

    if (!res.ok) {
      throw new Error(data.errors?.[0] || "Unexpected Error");
    }

    return data.content;
  },
});

// Mock function to handle row deletion
const mockDeleteRow = (rowId: string) => {
  console.log(`Delete row with ID: ${rowId}`);
  // In a real implementation, you would call an API endpoint here
  // and potentially use queryClient.invalidateQueries to refresh data
};

enum ColumnType {
  Integer = "integer",
  Serial = "serial",
  Varchar = "varchar",
  Text = "text",
  CharacterVarying = "character varying(255)",
  Boolean = "boolean",
  Date = "date",
  Timestamp = "timestamp",
  TimestampWithoutTimeZone = "timestamp without time zone",
  Float = "float",
  UUID = "uuid",
  JSON = "json",
}

interface ColumnIconProps {
  type: ColumnType | string;
}

const ColumnIcon: React.FC<ColumnIconProps> = ({ type }) => {
  let Icon = Circle;

  console.log("Column type:", type);

  switch (type) {
    case ColumnType.Integer:
      Icon = Hash;
      break;
    case ColumnType.Serial:
      Icon = Hash;
      break;
    case ColumnType.Varchar:
      Icon = Text;
      break;
    case ColumnType.CharacterVarying:
      Icon = Text;
      break;
    case ColumnType.Text:
      Icon = AlignLeft;
      break;
    case ColumnType.Boolean:
      Icon = ToggleLeft;
      break;
    case ColumnType.Date:
      Icon = Calendar;
      break;
    case ColumnType.Timestamp:
      Icon = Clock;
      break;
    case ColumnType.TimestampWithoutTimeZone:
      Icon = Clock;
      break;
    case ColumnType.Float:
      Icon = Hash;
      break;
    case ColumnType.UUID:
      Icon = Shuffle;
      break;
    case ColumnType.JSON:
      Icon = FileJson;
      break;
    default:
      Icon = Circle;
      break;
  }

  return <Icon className="mr-2 h-3 w-3 flex-shrink-0 translate-y-[0.5px]" />;
};

export const prepareColumns = (
  columns: any[] | undefined | null,
  collectionName?: string
): ColumnDef<any>[] => {
  if (!columns || !Array.isArray(columns) || columns.length === 0) {
    return [];
  }

  // Create the select column
  const selectColumn: ColumnDef<any, unknown> = {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
        className="ml-2"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
        className="ml-2"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  };

  // // Create the actions column with column visibility controls
  const ActionsColumnHeader = ({ table }: { table: any }) => {
    // State to control dropdown open/close
    const [dropdownOpen, setDropdownOpen] = useState(false);
    
    // Get column visibility state to trigger re-renders
    const columnVisibility = table?.getState?.().columnVisibility;
    
    // Calculate hidden columns count
    const hiddenColumnsCount = React.useMemo(() => {
      if (!table || !table.getAllColumns) return 0;
      return table.getAllColumns().filter((column: any) => 
        column && column.getCanHide && column.getCanHide() && !column.getIsVisible()
      ).length;
    }, [table, columnVisibility]);

    const hasHiddenColumns = hiddenColumnsCount > 0;

    return (
      <div className="flex items-center justify-end pr-6">
        <DropdownMenu
          open={dropdownOpen}
          onOpenChange={setDropdownOpen}
          modal={false}
        >
          <Tooltip>
            <TooltipTrigger asChild>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0 rounded-full hover:bg-muted data-[state=open]:bg-muted hover:shadow-sm relative"
                >
                  {hasHiddenColumns ? (
                    <>
                      <EyeOff className="h-4 w-4" />
                      <span className="absolute -top-1 -right-1 h-2 w-2 bg-primary rounded-full" />
                    </>
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                  <span className="sr-only">
                    {hasHiddenColumns 
                      ? `Column visibility options (${hiddenColumnsCount} hidden)` 
                      : "Column visibility options"}
                  </span>
                </Button>
              </DropdownMenuTrigger>
            </TooltipTrigger>
            <TooltipContent>
              <p>
                {hasHiddenColumns 
                  ? `${hiddenColumnsCount} column${hiddenColumnsCount > 1 ? 's' : ''} hidden` 
                  : "Show/hide columns"}
              </p>
            </TooltipContent>
          </Tooltip>
          <DropdownMenuContent
            align="end"
            className="w-[220px]"
            onCloseAutoFocus={(e) => {
              e.preventDefault();
              // Prevent closing when interacting with content
              e.stopPropagation();
            }}
            onEscapeKeyDown={() => setDropdownOpen(false)}
            forceMount
          >
            <div className="flex items-center justify-between px-2 py-1.5">
              <DropdownMenuLabel className="px-0 py-0">
                Column Visibility
              </DropdownMenuLabel>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 w-8 p-0 rounded-full hover:bg-muted"
                onClick={(e) => {
                  table.resetColumnVisibility();
                  e.stopPropagation();
                }}
                title="Reset all columns to visible"
              >
                <RotateCcw className="h-3.5 w-3.5" />
              </Button>
            </div>
            <DropdownMenuSeparator />
            {Array.isArray(table.getAllColumns())
              ? table.getAllColumns()
                  .filter((column: any) => column && column.getCanHide && column.getCanHide())
                  .map((column: any) => {
                    return (
                      <DropdownMenuItem
                        key={column.id}
                        className="flex items-center justify-between py-2 px-2"
                        onSelect={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                        }}
                      >
                        <div className="flex items-center gap-2">
                          <span>{column.id}</span>
                        </div>
                        <Switch
                          checked={column.getIsVisible()}
                          disabled={!column.getCanHide()}
                          onCheckedChange={column.toggleVisibility}
                        />
                      </DropdownMenuItem>
                    );
                  })
              : null}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    );
  };

  const actionsColumn: ColumnDef<any, unknown> = {
    id: "actions",
    accessorKey: "actions",
    header: ({ table }) => (
      <ActionsColumnHeader key="actions-column-header" table={table} />
    ),
    enableHiding: false,
    cell: ({ row, table }) => {
      const rowId = row.original?.id;

      return (
        <div className="flex justify-end pr-6">
          <DropdownMenu modal={false}>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 w-8 p-0 rounded-full hover:bg-muted data-[state=open]:bg-muted hover:shadow-sm"
              >
                <MoreVertical className="h-4 w-4" />
                <span className="sr-only">Open menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-[160px]">
              <DropdownMenuItem
                onClick={() =>
                  mockDeleteRow(String(row.original?.id || row.id))
                }
                className="text-destructive focus:text-destructive"
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      );
    },
    meta: {
      isSticky: true,
    },
  };

  // Create the data columns
  const dataColumns = Array.isArray(columns)
    ? columns.filter(Boolean).map((column, index) => {
        const dataType = column.type || "";
        const isTimestamp = dataType.includes("timestamp");
        const columnId = column.name;

        return {
          id: columnId,
          accessorKey: columnId,
          accessorFn: (row: any) =>
            row && column && column.name ? row[column.name] : null,
          header: () => (
            <div className="inline-flex items-center whitespace-nowrap">
              <ColumnIcon type={column.type} />
              <span className="truncate font-medium block max-w-[200px]">
                {column && column.name ? column.name : ""}
              </span>
            </div>
          ),
          cell: ({ row, table }: { row: Row<any>; table: any }) => {
            // Get value safely
            let value: CellValue = null;
            try {
              if (row && column && column.name) {
                value = row.getValue(column.name) as CellValue;
              }
            } catch (e) {
              // Handle error silently
            }
            
            return (
              <OptimisticTableCell
                value={value}
                column={column}
                rowId={row.original?.id}
                rowData={row.original}
                isTimestamp={isTimestamp}
                onUpdate={table.options.meta?.onCellUpdate}
              />
            );
          },
          meta: {
            collectionName,
          },
        };
      })
    : [];

  return [selectColumn, ...dataColumns, actionsColumn];
};

export const rowsQuery = (
  projectId: string,
  dbId: string,
  collectionName: string,
  pagination: PaginationState = { pageIndex: 0, pageSize: 50 },
  filters: Record<string, string> = {}
) => ({
  queryKey: [
    "rows",
    projectId,
    collectionName,
    pagination.pageSize,
    pagination.pageIndex,
    filters,
  ],
  queryFn: async () => {
    if (!collectionName) {
      return { rows: [], totalCount: 0 };
    }

    const offset = pagination.pageIndex * pagination.pageSize;
    const limit = pagination.pageSize;

    const authToken = await getClientAuthToken();

    if (!authToken) {
      throw new Error("No auth token");
    }

    const services = initializeServices(authToken);
    const projectToken = await services.projects.getProjectToken(projectId);

    try {
      const res = await services.tables.getTableRows(
        projectId,
        collectionName,
        {
          headers: {
            Prefer: "count=exact",
            Authorization: `Bearer ${projectToken}`,
          },
          params: {
            limit,
            offset,
            order: "updated_at.desc", // Sort by updated_at descending (newest first)
            ...filters,
          },
          baseUrl: getProjectApiUrl(dbId),
        }
      );

      if (!res.ok) {
        const responseData = await res.json();
        const errorMessage = responseData?.errors[0] || "Unknown error";

        if (res.status === 401) {
          // throw new UnauthorizedError(errorMessage);
        } else {
          throw new Error(errorMessage);
        }
      }

      // Extract total count from Content-Range header
      let totalCount = 0;
      const contentRange = res.headers.get("Content-Range");
      if (contentRange) {
        // Format is typically like "0-49/100" or "0-49/*"
        const match = contentRange.match(/\/(\d+|\*)/);
        if (match && match[1] && match[1] !== "*") {
          totalCount = parseInt(match[1], 10);
        }
      }

      const data = await res.json();
      return {
        rows: Array.isArray(data) ? data : [],
        totalCount,
      };
    } catch (error) {
      console.error("Error fetching rows:", error);
      return { rows: [], totalCount: 0 };
    }
  },
});
