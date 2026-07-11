import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "~/components/ui/sheet";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Textarea } from "~/components/ui/textarea";
import { Checkbox } from "~/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import { Calendar } from "~/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import { CalendarIcon } from "lucide-react";
import { format } from "date-fns";
import { cn, getProjectApiUrl } from "~/lib/utils";
import type { TablesService } from "~/services/tables";
import type { ProjectService } from "~/services/projects";

interface EditRowSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  row: Record<string, any>;
  columns: Array<{
    name: string;
    type: string;
    is_nullable?: string;
    column_default?: string | null;
  }>;
  tableId: string;
  projectId: string;
  dbId: string;
  services: { tables: TablesService; projects: ProjectService };
  onSuccess?: () => void;
}

export function EditRowSheet({
  open,
  onOpenChange,
  row,
  columns,
  tableId,
  projectId,
  dbId,
  services,
  onSuccess,
}: EditRowSheetProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [jsonErrors, setJsonErrors] = useState<Record<string, string>>({});
  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors },
  } = useForm({
    defaultValues: row,
  });

  // Reset form when row changes
  useEffect(() => {
    reset(row);
    setJsonErrors({});
  }, [row, reset]);

  const onSubmit = async (data: Record<string, any>) => {
    try {
      setIsSubmitting(true);

      // Check for JSON validation errors
      const hasJsonErrors = Object.values(jsonErrors).some(
        (error) => error !== ""
      );
      if (hasJsonErrors) {
        toast.error("Please fix JSON validation errors before submitting");
        setIsSubmitting(false);
        return;
      }

      // Remove id, created_at, and updated_at fields from the data to be updated
      const { id, created_at, updated_at, ...updateData } = data;

      // Validate row ID exists
      if (!row.id) {
        toast.error("Row ID is missing");
        return;
      }

      const projectToken = await services.projects.getProjectToken(projectId);

      const response = await services.tables.updateTableRow(
        projectId,
        tableId,
        row.id,
        updateData,
        {
          baseUrl: getProjectApiUrl(dbId),
          headers: { Authorization: `Bearer ${projectToken}` },
        }
      );

      if (response.ok) {
        toast.success("Row updated successfully");
        onOpenChange(false);
        onSuccess?.();
      } else {
        let errorMessage = "Failed to update row";
        try {
          const errorData = await response.json();
          errorMessage =
            errorData.message ||
            errorData.error ||
            errorData.errors?.[0] ||
            errorMessage;
        } catch {
          // If not JSON, try text
          try {
            const text = await response.text();
            if (text) errorMessage = text;
          } catch {
            // Use default error message
          }
        }
        toast.error(errorMessage);
      }
    } catch (error) {
      console.error("Error updating row:", error);
      toast.error("An error occurred while updating the row");
    } finally {
      setIsSubmitting(false);
    }
  };

  const renderField = (column: any) => {
    const fieldName = column.name;
    const dataType = (column.type || "").toLowerCase();
    const isNullable = column.is_nullable === "YES";
    const isIdField = fieldName === "id";
    const isTimestampField =
      fieldName === "created_at" || fieldName === "updated_at";

    // Common props for all inputs
    const commonProps = {
      ...register(fieldName),
      disabled: isIdField || isTimestampField || isSubmitting,
      className: cn(
        (isIdField || isTimestampField) && "opacity-50 cursor-not-allowed"
      ),
    };

    // Handle different data types
    if (dataType === "boolean" || dataType.includes("bool")) {
      return (
        <div className="flex items-center space-x-2">
          <Checkbox
            id={fieldName}
            checked={watch(fieldName) || false}
            onCheckedChange={(checked) => setValue(fieldName, checked)}
            disabled={isIdField || isTimestampField || isSubmitting}
          />
          <Label htmlFor={fieldName} className="text-sm font-normal">
            {isNullable ? "Optional" : "Required"}
          </Label>
        </div>
      );
    }

    if (dataType.includes("date") || dataType.includes("timestamp")) {
      const dateValue = watch(fieldName);
      return (
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className={cn(
                "w-full justify-start text-left font-normal",
                !dateValue && "text-muted-foreground",
                (isIdField || isTimestampField) &&
                  "opacity-50 cursor-not-allowed"
              )}
              disabled={isIdField || isTimestampField || isSubmitting}
            >
              <CalendarIcon className="mr-2 h-4 w-4" />
              {dateValue ? format(new Date(dateValue), "PPP") : "Pick a date"}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0" align="start">
            <Calendar
              mode="single"
              selected={dateValue ? new Date(dateValue) : undefined}
              onSelect={(date) => setValue(fieldName, date?.toISOString())}
            />
          </PopoverContent>
        </Popover>
      );
    }

    if (dataType.includes("text")) {
      return (
        <Textarea
          {...commonProps}
          placeholder={isNullable ? "Optional" : "Required"}
          rows={3}
        />
      );
    }

    if (dataType.includes("json")) {
      return (
        <div>
          <Textarea
            {...commonProps}
            placeholder={isNullable ? "Optional JSON" : "Required JSON"}
            rows={4}
            onChange={(e) => {
              const value = e.target.value;
              if (value) {
                try {
                  JSON.parse(value); // Validate JSON
                  setJsonErrors((prev) => ({ ...prev, [fieldName]: "" }));
                } catch {
                  setJsonErrors((prev) => ({
                    ...prev,
                    [fieldName]: "Invalid JSON format",
                  }));
                }
              } else {
                setJsonErrors((prev) => ({ ...prev, [fieldName]: "" }));
              }
              register(fieldName).onChange(e);
            }}
          />
          {jsonErrors[fieldName] && (
            <span className="text-sm text-destructive mt-1 block">
              {jsonErrors[fieldName]}
            </span>
          )}
        </div>
      );
    }

    // Default to input for numbers, strings, etc.
    return (
      <Input
        {...commonProps}
        type={
          dataType.includes("int") ||
          dataType.includes("serial") ||
          dataType.includes("float") ||
          dataType.includes("numeric")
            ? "number"
            : "text"
        }
        placeholder={isNullable ? "Optional" : "Required"}
      />
    );
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-[540px] overflow-y-auto">
        <SheetHeader className="px-6">
          <SheetTitle>Edit Row</SheetTitle>
          <SheetDescription>
            Make changes to the row data. The ID field cannot be modified.
          </SheetDescription>
        </SheetHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 mt-6 px-6">
          {columns.map((column, index) => (
            <div key={column.name} className="space-y-2">
              <Label htmlFor={column.name}>
                {column.name}
                {(column.name === "id" ||
                  column.name === "created_at" ||
                  column.name === "updated_at") && (
                  <span className="ml-2 text-xs text-muted-foreground">
                    (Read-only)
                  </span>
                )}
              </Label>
              {renderField(column)}
            </div>
          ))}
        </form>
        <div className="flex justify-end gap-2 mt-6 px-6 pb-6">
          <Button
            type="button"
            variant="secondary"
            onClick={() => onOpenChange(false)}
            disabled={isSubmitting}
            className="cursor-pointer"
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={isSubmitting}
            onClick={handleSubmit(onSubmit)}
            className="cursor-pointer"
          >
            {isSubmitting ? "Updating..." : "Update"}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}

