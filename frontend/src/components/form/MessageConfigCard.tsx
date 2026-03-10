"use client";

import { UseFormReturn } from "react-hook-form";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FileTextIcon } from "lucide-react";
import { TemplateManager } from "@/components/template-manager";
import { FormData } from "./types";
import { FACILITY_OPTIONS, SEVERITY_OPTIONS } from "./constants";

// ================================================================================
// MESSAGE CONFIG CARD COMPONENT
// Single Responsibility: Handles message configuration UI (RFC format, facility, etc.)
// ================================================================================

interface MessageConfigCardProps {
  form: UseFormReturn<FormData>;
}

export function MessageConfigCard({ form }: MessageConfigCardProps) {
  return (
    <Card>
      <CardHeader className="pb-2 pt-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <FileTextIcon className="w-4 h-4" />
            Message Configuration
          </CardTitle>
          <TemplateManager
            currentValues={{
              Messages: form.watch("Messages"),
              Facility: form.watch("Facility"),
              Severity: form.watch("Severity"),
              Appname: form.watch("Appname"),
              UseRFC5424: form.watch("UseRFC5424"),
            }}
            onLoadTemplate={(template) => {
              form.setValue("Messages", template.message);
              form.setValue("Facility", template.facility);
              form.setValue("Severity", template.severity);
              form.setValue("Appname", template.appname);
              form.setValue("UseRFC5424", template.useRfc5424);
            }}
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-3 pb-3">
        {/* RFC Format Toggle */}
        <FormField
          control={form.control}
          name="UseRFC5424"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center gap-3 space-y-0 rounded-md border border-border/60 p-3 bg-secondary/20">
              <FormControl>
                <input
                  type="checkbox"
                  checked={field.value}
                  onChange={field.onChange}
                  className="h-4 w-4 rounded accent-primary"
                />
              </FormControl>
              <div className="flex-1 space-y-0.5">
                <FormLabel className="text-xs font-medium leading-none">
                  Use RFC 5424 Format
                </FormLabel>
                <FormDescription className="text-[10px] text-muted-foreground">
                  Modern format (checked) or legacy RFC 3164
                </FormDescription>
              </div>
            </FormItem>
          )}
        />

        {/* Facility & Severity */}
        <div className="grid grid-cols-2 gap-3">
          <FormField
            control={form.control}
            name="Facility"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="text-xs font-medium">Facility</FormLabel>
                <Select
                  onValueChange={(value: string) => field.onChange(parseInt(value))}
                  defaultValue={field.value?.toString()}
                >
                  <FormControl>
                    <SelectTrigger className="h-9">
                      <SelectValue placeholder="Select" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent className="max-h-[200px]">
                    {FACILITY_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value.toString()}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="Severity"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="text-xs font-medium">Severity</FormLabel>
                <Select
                  onValueChange={(value: string) => field.onChange(parseInt(value))}
                  defaultValue={field.value?.toString()}
                >
                  <FormControl>
                    <SelectTrigger className="h-9">
                      <SelectValue placeholder="Select" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent className="max-h-[200px]">
                    {SEVERITY_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value.toString()}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />
        </div>

        {/* Hostname & App Name */}
        <div className="grid grid-cols-2 gap-3">
          <FormField
            control={form.control}
            name="Hostname"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="text-xs font-medium">Hostname</FormLabel>
                <FormControl>
                  <Input placeholder="Optional" className="h-9" {...field} />
                </FormControl>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="Appname"
            render={({ field }) => (
              <FormItem>
                <FormLabel className="text-xs font-medium">App Name</FormLabel>
                <FormControl>
                  <Input placeholder="sendlog" className="h-9" {...field} />
                </FormControl>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />
        </div>
      </CardContent>
    </Card>
  );
}
