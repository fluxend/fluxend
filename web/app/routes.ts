import {
  type RouteConfig,
  index,
  layout,
  prefix,
  route,
} from "@react-router/dev/routes";
import { routeFolder } from "./lib/router";

export default [
  index("./routes/auth/login.tsx"),
  route("signup", "./routes/auth/signup.tsx"),
  route("logout", "./routes/auth/logout.tsx"),
  layout("./components/shared/app-layout.tsx", [
    route("/settings", "./routes/settings/page.tsx"),
    route("projects", "./routes/projects/page.tsx"),
    route("projects/create", "./routes/projects/create.tsx"),
  ]),
  ...prefix("projects/:projectId", [
    layout("./components/shared/project-layout.tsx", [
      route("dashboard", "./routes/dashboard/page.tsx"),
      route("docs", "./routes/docs/sidebar.tsx", [
        index("./routes/docs/overview.tsx"),
        route(":table", "./routes/docs/table-docs.tsx"),
      ]),
      route("tables/create", "./routes/tables/create.tsx"),
      route("tables/:tableId/edit", "./routes/tables/edit.tsx"),
      route("tables/:tableId/webhooks", "./routes/tables/webhooks.tsx"),
      route("tables", "./routes/tables/sidebar.tsx", [
        route(":tableId", "./routes/tables/page.tsx"),
      ]),
      routeFolder("functions", "./routes/functions/"),
      routeFolder("logs", "./routes/logs/"),
      route("storage", "./routes/storage/sidebar.tsx", [
        route(":containerId", "./routes/storage/$containerId.tsx"),
      ]),
      routeFolder("forms", "./routes/forms/"),
      routeFolder("backups", "./routes/backups/"),
    ]),
  ]),
] satisfies RouteConfig;