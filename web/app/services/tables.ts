import type { APIResponse } from "~/lib/types";
import { getTypedResponseData } from "~/lib/utils";
import type { Table } from "~/routes/tables/table-list";
import {get, post, del, put, type APIRequestOptions, patch} from "~/tools/fetch";

export function createTablesService(authToken: string) {
  const getAllTables = async (projectId: string) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    const response = await get("/tables", fetchOptions);
    const data = await getTypedResponseData<APIResponse<any>>(response);

    return data;
  };

  const getTableColumns = async (projectId: string, collectionName: string) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return get(`/tables/public.${collectionName}/columns`, fetchOptions);
  };

  const getTableRows = async (
    projectId: string,
    collectionName: string,
    options?: {
      headers?: HeadersInit;
      params?: Record<string, any>;
      baseUrl?: string;
    }
  ) => {
    const fetchOptions: APIRequestOptions = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
        ...options?.headers,
      },
      params: options?.params,
      baseUrl: options?.baseUrl,
    };

    return get(collectionName, fetchOptions);
  };

  const createTable = async (projectId: string, data: any) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return post("/tables", data, fetchOptions);
  };

  const deleteTable = async (projectId: string, tableName: string) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    const response = await del(`/tables/public.${tableName}`, fetchOptions);
    const data = await getTypedResponseData<APIResponse<null>>(response);
    return data;
  };

  const createTableColumns = async (projectId: string, tableName: string, data: any) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return post(`/tables/public.${tableName}/columns`, data, fetchOptions);
  };

  const updateTableColumns = async (projectId: string, tableName: string, data: any) => {
    const fetchOptions: RequestInit = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
    };

    return patch(`/tables/public.${tableName}/columns`, data, fetchOptions);
  };

  const updateTableRow = async (
    projectId: string,
    tableId: string,
    rowId: string,
    data: any,
    options?: {
      headers?: HeadersInit;
      baseUrl?: string;
    }
  ) => {
    const fetchOptions: APIRequestOptions = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
        "Prefer": "return=representation", // Return the updated row
        ...options?.headers,
      },
      baseUrl: options?.baseUrl,
      params: {
        id: `eq.${rowId}` // PostgREST filter syntax
      }
    };

    return patch(tableId, data, fetchOptions);
  };

  const deleteTableRows = async (
    projectId: string,
    tableId: string,
    rowIds: string[],
    options?: {
      headers?: HeadersInit;
      baseUrl?: string;
    }
  ) => {
    // Validate input
    if (!rowIds || rowIds.length === 0) {
      throw new Error('No row IDs provided for deletion');
    }

    // Escape IDs to prevent injection
    const escapedIds = rowIds.map(id => 
      // Remove any special characters that could break the query
      String(id).replace(/[^a-zA-Z0-9-_]/g, '')
    ).filter(Boolean);

    if (escapedIds.length === 0) {
      throw new Error('No valid row IDs provided for deletion');
    }

    const fetchOptions: APIRequestOptions = {
      headers: {
        "X-Project": projectId,
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
        "Prefer": "count=exact", // Return count of deleted rows
        ...options?.headers,
      },
      baseUrl: options?.baseUrl,
      params: {
        id: `in.(${escapedIds.join(",")})` // PostgREST filter syntax for multiple IDs
      }
    };

    return del(tableId, fetchOptions);
  };

  return {
    getAllTables,
    getTableColumns,
    getTableRows,
    createTable,
    deleteTable,
    createTableColumns,
    updateTableColumns,
    updateTableRow,
    deleteTableRows,
  };
}

export type TablesService = ReturnType<typeof createTablesService>;
