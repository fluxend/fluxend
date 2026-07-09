import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import type { Route } from "./+types/page";
import {
  CheckCircleIcon,
  XCircleIcon,
  Database,
  Server,
  HardDrive,
  Cpu,
  TableIcon,
  BarChart3,
  Target,
  TrendingUp,
} from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { AppHeader } from "~/components/shared/header";
import { redirect, useOutletContext } from "react-router";
import type { ProjectLayoutOutletContext } from "~/components/shared/project-layout";
import { getAuthToken, getClientAuthToken } from "~/lib/auth";
import { getProjectApiUrl } from "~/lib/utils";
import { initializeServices } from "~/services";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Dashboard - Fluxend" },
    { name: "description", content: "Fluxend Dashboard" },
  ];
}

function StatusCard({
    title,
    status,
    subtitle,
    icon: Icon,
    isStale,
  }: {
  title: string;
  status: string;
  subtitle?: string;
  icon: React.ComponentType<{ className?: string }>;
  isStale: boolean;
}) {
  return (
      <Card className="flex-1 min-w-[120px] rounded-lg">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <Icon className="size-4 text-muted-foreground" />
            {isStale ? (
                <XCircleIcon className="size-4 text-red-500" />
            ) : (
                <CheckCircleIcon className="size-4 text-green-500" />
            )}
          </div>
          <CardDescription className="text-xs">{title}</CardDescription>
          <CardTitle className="text-lg font-semibold">{status}</CardTitle>
          {subtitle && (
              <div className="text-xs text-muted-foreground mt-1">{subtitle}</div>
          )}
        </CardHeader>
      </Card>
  );
}

function MetricCard({
    title,
    value,
    subtitle,
    icon: Icon,
    data,
    isStale,
  }: {
  title: string;
  value: string;
  subtitle?: string;
  icon: React.ComponentType<{ className?: string }>;
  data: number[];
  isStale: boolean;
}) {
  const maxValue = Math.max(...data);
  const minValue = Math.min(...data);

  return (
      <Card className="flex-1 min-w-[200px] rounded-lg">
        <CardHeader className="relative pb-2">
          <div className="flex items-center justify-between">
            <Icon className="size-4 text-muted-foreground" />
            {isStale ? (
                <XCircleIcon className="size-4 text-red-500" />
            ) : (
                <span className="relative flex size-3">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75"></span>
              <span className="relative inline-flex size-3 rounded-full bg-green-500"></span>
            </span>
            )}
          </div>
          <CardDescription className="text-xs">{title}</CardDescription>
          <CardTitle className="text-2xl font-semibold tabular-nums">
            {value}
          </CardTitle>
        </CardHeader>
        <CardFooter className="pt-0">
          <div className="w-full">
            {subtitle && (
                <div className="text-xs text-muted-foreground">{subtitle}</div>
            )}
          </div>
        </CardFooter>
      </Card>
  );
}

function StatsCard({
     title,
     value,
     subtitle,
     icon: Icon,
     isStale,
   }: {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ComponentType<{ className?: string }>;
  isStale: boolean;
}) {
  return (
      <Card className="flex-1 min-w-[180px] rounded-lg">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <Icon className="size-4 text-muted-foreground" />
            {isStale ? (
                <XCircleIcon className="size-4 text-red-500" />
            ) : (
                <CheckCircleIcon className="size-4 text-green-500" />
            )}
          </div>
          <CardDescription className="text-xs">{title}</CardDescription>
          <CardTitle className="text-xl font-semibold tabular-nums">
            {value}
          </CardTitle>
          {subtitle && (
              <div className="text-xs text-muted-foreground mt-1">{subtitle}</div>
          )}
        </CardHeader>
      </Card>
  );
}

export async function clientLoader({ params }: Route.LoaderArgs) {
  const authToken = await getClientAuthToken();
  if (!authToken) {
    return redirect("/logout");
  }

  const services = initializeServices(authToken);
  const { projectId } = params;

  const [healthData, projectStats] = await Promise.allSettled([
    services.dashboard.getHealthStatus(),
    services.dashboard.getProjectStats(projectId),
  ]);

  return {
    healthData: healthData.status === "fulfilled" ? healthData.value : null,
    projectStats: projectStats.status === "fulfilled" ? projectStats.value : null,
  };
}

export default function Dashboard({ loaderData }: Route.ComponentProps) {
  const { services, project } = useOutletContext<ProjectLayoutOutletContext>();

  const {
    isLoading: isHealthLoading,
    isError: isHealthError,
    data: healthData,
    isStale: isHealthStale,
  } = useQuery({
    queryKey: ["dashboard-health"],
    initialData: loaderData.healthData,
    queryFn: async () => {
      return await services.dashboard.getHealthStatus();
    },
    staleTime: 15000,
    refetchInterval: 10000,
    refetchIntervalInBackground: true,
  });

  const {
    isLoading: isStatsLoading,
    isError: isStatsError,
    data: projectStats,
    isStale: isStatsStale,
  } = useQuery({
    queryKey: ["project-stats", project?.uuid],
    initialData: loaderData.projectStats,
    queryFn: async () => {
      if (!project?.uuid) return null;
      return await services.dashboard.getProjectStats(project.uuid);
    },
    staleTime: 30000,
    refetchInterval: 30000,
    refetchIntervalInBackground: true,
    enabled: !!project?.uuid,
  });

  // Generate mock historical data for charts
  const diskData = useMemo(() => {
    const baseValue = parseFloat(
        healthData?.disk_usage?.replace("%", "") || "48.9"
    );
    return Array.from({ length: 8 }, (_, i) =>
        Math.max(0, Math.min(100, baseValue + (Math.random() - 0.5) * 10))
    );
  }, [healthData?.disk_usage]);

  const cpuData = useMemo(() => {
    const baseValue = parseFloat(
        healthData?.cpu_usage?.replace("%", "") || "0"
    );
    return Array.from({ length: 8 }, (_, i) =>
        Math.max(0, Math.min(100, baseValue + (Math.random() - 0.5) * 5))
    );
  }, [healthData?.cpu_usage]);

  // Calculate project stats
  const projectMetrics = useMemo(() => {
    if (!projectStats || !projectStats.tableCount || !projectStats.tableSize) return null;

    const totalTables = projectStats.tableCount.length;

    // Handle empty arrays
    if (projectStats.tableCount.length === 0 || projectStats.tableSize.length === 0) {
      return {
        totalTables: 0,
        totalSize: projectStats.totalSize || "0 kB",
        indexSize: projectStats.indexSize || "0 kB",
        tableWithMostRows: "No tables",
        tableWithBiggestSize: "No tables",
      };
    }

    const tableWithMostRows = projectStats.tableCount.reduce((max, current) =>
        (current?.EstimatedRowCount || 0) > (max?.EstimatedRowCount || 0) ? current : max
    );

    const tableWithBiggestSize = projectStats.tableSize.reduce((max, current) => {
      const currentSize = parseFloat((current?.totalSize || "0").replace(/[^\d.]/g, '')) || 0;
      const maxSize = parseFloat((max?.totalSize || "0").replace(/[^\d.]/g, '')) || 0;
      return currentSize > maxSize ? current : max;
    });

    return {
      totalTables,
      totalSize: projectStats.totalSize,
      indexSize: projectStats.indexSize,
      tableWithMostRows: `${tableWithMostRows.TableName} (${tableWithMostRows.EstimatedRowCount.toLocaleString()} rows)`,
      tableWithBiggestSize: `${tableWithBiggestSize.tableName} (${tableWithBiggestSize.totalSize})`,
    };
  }, [projectStats]);

  if (isHealthLoading && isStatsLoading) {
    return (
        <>
          <AppHeader title="Dashboard" />
          <div className="flex flex-col gap-4 p-4">
            <div className="text-center">Loading dashboard...</div>
          </div>
        </>
    );
  }

  return (
      <>
        <AppHeader title="Dashboard" />
        <div className="flex flex-col gap-8 p-4">
          {/* System Health Section */}
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Server className="size-5 text-muted-foreground" />
              <h2 className="text-lg font-semibold">System Health</h2>
            </div>

            {isHealthError || !healthData ? (
                <div className="text-center text-red-500 py-8">
                  Error loading system health data
                </div>
            ) : (
                <>
                  <div className="flex flex-wrap gap-4">
                    <StatusCard
                        title="Database"
                        status={healthData.database_status}
                        subtitle={projectStats?.databaseName || "Loading..."}
                        icon={Database}
                        isStale={!isHealthLoading && isHealthStale}
                    />
                    <StatusCard
                        title="UI App"
                        status={healthData.app_status}
                        icon={Server}
                        isStale={!isHealthLoading && isHealthStale}
                    />
                    <StatusCard
                        title="PostgREST"
                        status={healthData.postgrest_status}
                        subtitle={projectStats?.databaseName ? getProjectApiUrl(projectStats.databaseName) : "Loading..."}
                        icon={Server}
                        isStale={!isHealthLoading && isHealthStale}
                    />
                  </div>

                  {/* Metrics Cards Row */}
                  <div className="flex flex-wrap gap-4">
                    <MetricCard
                        title="Disk Usage"
                        value={healthData.disk_usage}
                        subtitle={`${healthData.disk_available} available of ${healthData.disk_total}`}
                        icon={HardDrive}
                        data={diskData}
                        isStale={isHealthStale}
                    />
                    <MetricCard
                        title="CPU Usage"
                        value={healthData.cpu_usage}
                        subtitle={`${healthData.cpu_cores} core${
                            healthData.cpu_cores !== 1 ? "s" : ""
                        } • Real-time utilization`}
                        icon={Cpu}
                        data={cpuData}
                        isStale={isHealthStale}
                    />
                  </div>
                </>
            )}
          </div>

          {/* Project Stats Section */}
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <BarChart3 className="size-5 text-muted-foreground" />
              <h2 className="text-lg font-semibold">Project Statistics</h2>
            </div>

            {isStatsError || !projectStats || !projectMetrics ? (
                <div className="text-center text-muted-foreground py-8">
                  {isStatsError ? "Error loading project statistics" : "No project data available"}
                </div>
            ) : (
                <div className="flex flex-wrap gap-4">
                  <StatsCard
                      title="Total Tables"
                      value={projectMetrics.totalTables}
                      icon={TableIcon}
                      isStale={!isStatsLoading && isStatsStale}
                  />
                  <StatsCard
                      title="Total Size"
                      value={projectMetrics.totalSize}
                      subtitle="All tables combined"
                      icon={Database}
                      isStale={!isStatsLoading && isStatsStale}
                  />
                  <StatsCard
                      title="Index Size"
                      value={projectMetrics.indexSize}
                      subtitle="All indexes combined"
                      icon={Target}
                      isStale={!isStatsLoading && isStatsStale}
                  />
                  <StatsCard
                      title="Most Rows"
                      value={projectMetrics.tableWithMostRows}
                      icon={TrendingUp}
                      isStale={!isStatsLoading && isStatsStale}
                  />
                </div>
            )}
          </div>
        </div>
      </>
  );
}