import { useState } from "react";
import { useParams, useOutletContext } from "react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Trash2, Webhook, Plus } from "lucide-react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import type { ProjectLayoutOutletContext } from "~/components/shared/project-layout";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Badge } from "~/components/ui/badge";
import { Checkbox } from "~/components/ui/checkbox";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "~/components/ui/form";
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
import type { APIResponse } from "~/lib/types";
import type { Webhook as WebhookType } from "~/services/webhooks";

const EVENT_OPTIONS = [
  { value: "insert", label: "Insert" },
  { value: "update", label: "Update" },
  { value: "delete", label: "Delete" },
] as const;

const createWebhookSchema = z.object({
  url: z.string().url("Must be a valid URL"),
  events: z.array(z.string()).min(1, "Select at least one event"),
  is_active: z.boolean(),
});

type CreateWebhookForm = z.infer<typeof createWebhookSchema>;

export function meta() {
  return [{ title: "Webhooks - Fluxend" }];
}

export default function WebhooksPage() {
  const { projectId, tableId } = useParams();
  const { services } = useOutletContext<ProjectLayoutOutletContext>();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);

  const form = useForm<CreateWebhookForm>({
    resolver: zodResolver(createWebhookSchema),
    defaultValues: { url: "", events: [], is_active: true },
  });

  const { data: webhooks = [], isLoading } = useQuery<WebhookType[]>({
    queryKey: ["webhooks", projectId, tableId],
    queryFn: async () => {
      const response = await services.webhooks.listWebhooks(projectId!, tableId!);
      const data = (await response.json()) as APIResponse<WebhookType[]>;
      return data.content ?? [];
    },
    enabled: !!projectId && !!tableId,
  });

  const createMutation = useMutation({
    mutationFn: async (values: CreateWebhookForm) => {
      const response = await services.webhooks.createWebhook(
        projectId!,
        tableId!,
        values
      );
      const data = (await response.json()) as APIResponse<WebhookType>;
      if (!response.ok) {
        throw new Error(data.errors?.[0] ?? "Failed to create webhook");
      }
      return data.content;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks", projectId, tableId] });
      toast.success("Webhook created");
      form.reset();
      setShowForm(false);
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const deleteMutation = useMutation({
    mutationFn: async (webhookUUID: string) => {
      const response = await services.webhooks.deleteWebhook(
        projectId!,
        tableId!,
        webhookUUID
      );
      if (!response.ok) {
        throw new Error("Failed to delete webhook");
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks", projectId, tableId] });
      toast.success("Webhook deleted");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const onSubmit = (values: CreateWebhookForm) => createMutation.mutate(values);

  return (
    <div className="flex flex-col h-full overflow-auto">
      <div className="border-b px-4 py-2 mb-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Webhook className="size-4" />
            <span className="text-base font-bold">
              Webhooks / {tableId}
            </span>
          </div>
          <Button size="sm" onClick={() => setShowForm((v) => !v)}>
            <Plus className="size-4 mr-1" />
            Add Webhook
          </Button>
        </div>
      </div>

      <div className="px-4 space-y-4">
        {showForm && (
          <div className="border rounded-lg p-4 bg-muted/10">
            <h3 className="text-sm font-semibold mb-3">New Webhook</h3>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="url"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Endpoint URL</FormLabel>
                      <FormControl>
                        <Input placeholder="https://example.com/webhook" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="events"
                  render={() => (
                    <FormItem>
                      <FormLabel>Events</FormLabel>
                      <div className="flex gap-4">
                        {EVENT_OPTIONS.map((opt) => (
                          <FormField
                            key={opt.value}
                            control={form.control}
                            name="events"
                            render={({ field }) => (
                              <FormItem className="flex items-center gap-2">
                                <FormControl>
                                  <Checkbox
                                    checked={field.value.includes(opt.value)}
                                    onCheckedChange={(checked) => {
                                      if (checked) {
                                        field.onChange([...field.value, opt.value]);
                                      } else {
                                        field.onChange(
                                          field.value.filter((v) => v !== opt.value)
                                        );
                                      }
                                    }}
                                  />
                                </FormControl>
                                <FormLabel className="font-normal cursor-pointer">
                                  {opt.label}
                                </FormLabel>
                              </FormItem>
                            )}
                          />
                        ))}
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="flex gap-2">
                  <Button
                    type="submit"
                    size="sm"
                    disabled={createMutation.isPending}
                  >
                    {createMutation.isPending ? "Saving..." : "Save"}
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => { setShowForm(false); form.reset(); }}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            </Form>
          </div>
        )}

        {isLoading && (
          <div className="text-sm text-muted-foreground">Loading webhooks...</div>
        )}

        {!isLoading && webhooks.length === 0 && !showForm && (
          <div className="flex items-center justify-center border rounded-lg p-8 bg-muted/10">
            <div className="text-center text-muted-foreground">
              <Webhook className="size-8 mx-auto mb-2 opacity-40" />
              <p className="text-sm">No webhooks configured for this table.</p>
            </div>
          </div>
        )}

        {webhooks.map((wh) => (
          <div
            key={wh.uuid}
            className="flex items-center justify-between border rounded-lg px-4 py-3 bg-background"
          >
            <div className="flex flex-col gap-1 min-w-0">
              <span className="text-sm font-medium truncate">{wh.url}</span>
              <div className="flex gap-1 flex-wrap">
                {wh.events.map((ev) => (
                  <Badge key={ev} variant="secondary" className="text-xs capitalize">
                    {ev}
                  </Badge>
                ))}
                {!wh.isActive && (
                  <Badge variant="outline" className="text-xs text-muted-foreground">
                    inactive
                  </Badge>
                )}
              </div>
            </div>

            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="ghost" size="icon" className="shrink-0 ml-2">
                  <Trash2 className="size-4" />
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete webhook?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will permanently remove the webhook endpoint{" "}
                    <strong>{wh.url}</strong>. No further deliveries will be made.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={() => deleteMutation.mutate(wh.uuid)}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        ))}
      </div>
    </div>
  );
}
