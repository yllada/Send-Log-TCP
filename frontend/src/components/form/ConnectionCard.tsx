"use client";

import { UseFormReturn } from "react-hook-form";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
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
import {
  NetworkIcon,
  CheckCircleIcon,
  XCircleIcon,
  FileIcon,
  FolderOpenIcon,
  XIcon,
  KeyIcon,
  ShieldCheckIcon,
} from "lucide-react";
import { ProfileManager } from "@/components/profile-manager";
import { FormData } from "./types";
import { DEFAULT_SYSLOG_PORT, DEFAULT_SYSLOG_TLS_PORT } from "./constants";
import { UseCertificatesReturn } from "./hooks/useCertificates";

// ================================================================================
// CONNECTION CARD COMPONENT
// Single Responsibility: Handles connection settings UI
// ================================================================================

interface ConnectionCardProps {
  form: UseFormReturn<FormData>;
  isConnected: boolean;
  onConnectionToggle: () => Promise<void>;
  certificates: UseCertificatesReturn;
}

export function ConnectionCard({
  form,
  isConnected,
  onConnectionToggle,
  certificates,
}: ConnectionCardProps) {
  const {
    handleSelectCACert,
    handleClearCACert,
    handleSelectClientCert,
    handleClearClientCert,
    handleSelectClientKey,
    handleClearClientKey,
  } = certificates;

  return (
    <Card>
      <CardHeader className="pb-2 pt-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <NetworkIcon className="w-4 h-4" />
            Connection Settings
          </CardTitle>
          <ProfileManager
            currentValues={{
              Address: form.watch("Address"),
              Port: form.watch("Port"),
              Protocol: form.watch("Protocol"),
              FramingMethod: form.watch("FramingMethod"),
              UseTLS: form.watch("UseTLS"),
              TLSVerify: form.watch("TLSVerify"),
              CACertPath: form.watch("CACertPath") || "",
              ClientCertPath: form.watch("ClientCertPath") || "",
              ClientKeyPath: form.watch("ClientKeyPath") || "",
            }}
            onLoadProfile={(profile) => {
              form.setValue("Address", profile.address);
              form.setValue("Port", profile.port);
              form.setValue("Protocol", profile.protocol);
              form.setValue("FramingMethod", profile.framingMethod as "octet-counting" | "non-transparent");
              form.setValue("UseTLS", profile.useTls);
              form.setValue("TLSVerify", profile.tlsVerify);
              form.setValue("CACertPath", profile.caCertPath || "");
              form.setValue("ClientCertPath", profile.clientCertPath || "");
              form.setValue("ClientKeyPath", profile.clientKeyPath || "");
            }}
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-3 pb-3">
        {/* Row 1: IP, Port, Protocol, Connect Button */}
        <div className="grid grid-cols-12 gap-3">
          <FormField
            control={form.control}
            name="Address"
            render={({ field }) => (
              <FormItem className="col-span-5 space-y-1">
                <FormLabel className="text-xs font-medium">IP Address</FormLabel>
                <FormControl>
                  <Input placeholder="192.168.1.100" className="h-9" {...field} />
                </FormControl>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="Port"
            render={({ field }) => (
              <FormItem className="col-span-2 space-y-1">
                <FormLabel className="text-xs font-medium">Port</FormLabel>
                <FormControl>
                  <Input placeholder="514" type="number" className="h-9" {...field} />
                </FormControl>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="Protocol"
            render={({ field }) => (
              <FormItem className="col-span-4 space-y-1">
                <FormLabel className="text-xs font-medium">Protocol</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger className="h-9">
                      <SelectValue placeholder="Select protocol" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="tcp">TCP</SelectItem>
                    <SelectItem value="udp">UDP</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage className="text-xs" />
              </FormItem>
            )}
          />

          <div className="col-span-1 flex items-end pb-[3px]">
            <Button
              type="button"
              onClick={onConnectionToggle}
              size="icon"
              className={`h-9 w-9 rounded-full transition-all ${
                isConnected
                  ? "bg-red-500 hover:bg-red-600 shadow-red-500/25"
                  : "bg-green-500 hover:bg-green-600 shadow-green-500/25"
              } shadow-lg`}
            >
              {isConnected ? (
                <XCircleIcon className="w-4 h-4" />
              ) : (
                <CheckCircleIcon className="w-4 h-4" />
              )}
            </Button>
          </div>
        </div>

        {/* Row 2: TCP Framing */}
        <FormField
          control={form.control}
          name="FramingMethod"
          render={({ field }) => {
            const isTCP = form.watch("Protocol") === "tcp";
            return (
              <FormItem>
                <FormLabel className={`text-xs font-medium ${!isTCP ? "text-muted-foreground" : ""}`}>
                  TCP Framing
                </FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value} disabled={!isTCP}>
                  <FormControl>
                    <SelectTrigger className="h-9" disabled={!isTCP}>
                      <SelectValue placeholder="Select method" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="octet-counting">Octet Counting (RFC 6587)</SelectItem>
                    <SelectItem value="non-transparent">Non-Transparent (LF)</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage className="text-xs" />
              </FormItem>
            );
          }}
        />

        {/* Row 3: TLS Options */}
        <div className="grid grid-cols-2 gap-3">
          <FormField
            control={form.control}
            name="UseTLS"
            render={({ field }) => {
              const isTCP = form.watch("Protocol") === "tcp";
              return (
                <FormItem className="flex flex-row items-center gap-3 space-y-0 rounded-md border border-border/60 p-3 bg-secondary/20">
                  <FormControl>
                    <input
                      type="checkbox"
                      checked={field.value}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        field.onChange(e.target.checked);
                        // Auto-adjust port when TLS is toggled
                        if (e.target.checked && form.getValues("Port") === DEFAULT_SYSLOG_PORT) {
                          form.setValue("Port", DEFAULT_SYSLOG_TLS_PORT);
                        } else if (!e.target.checked && form.getValues("Port") === DEFAULT_SYSLOG_TLS_PORT) {
                          form.setValue("Port", DEFAULT_SYSLOG_PORT);
                        }
                      }}
                      disabled={!isTCP}
                      className="h-4 w-4 rounded accent-primary"
                    />
                  </FormControl>
                  <div className="flex-1 space-y-0.5">
                    <FormLabel className={`text-xs font-medium leading-none ${!isTCP ? "text-muted-foreground" : ""}`}>
                      Use TLS/SSL
                    </FormLabel>
                    <FormDescription className="text-[10px] text-muted-foreground">
                      Encrypt connection (RFC 5425)
                    </FormDescription>
                  </div>
                </FormItem>
              );
            }}
          />

          <FormField
            control={form.control}
            name="TLSVerify"
            render={({ field }) => {
              const useTLS = form.watch("UseTLS");
              const isTCP = form.watch("Protocol") === "tcp";
              return (
                <FormItem className="flex flex-row items-center gap-3 space-y-0 rounded-md border border-border/60 p-3 bg-secondary/20">
                  <FormControl>
                    <input
                      type="checkbox"
                      checked={field.value}
                      onChange={field.onChange}
                      disabled={!isTCP || !useTLS}
                      className="h-4 w-4 rounded accent-primary"
                    />
                  </FormControl>
                  <div className="flex-1 space-y-0.5">
                    <FormLabel className={`text-xs font-medium leading-none ${!isTCP || !useTLS ? "text-muted-foreground" : ""}`}>
                      Verify Certificate
                    </FormLabel>
                    <FormDescription className="text-[10px] text-muted-foreground">
                      Uncheck for self-signed certs
                    </FormDescription>
                  </div>
                </FormItem>
              );
            }}
          />
        </div>

        {/* CA Certificate Selector */}
        <CertificateField
          form={form}
          name="CACertPath"
          label="CA Certificate (Optional)"
          icon={<FileIcon className="w-4 h-4 text-green-500 flex-shrink-0" />}
          emptyText="Using system CA store"
          description="Load a custom CA certificate (.pem, .crt) for server verification"
          onSelect={handleSelectCACert}
          onClear={handleClearCACert}
          showWhen={() => {
            const isTCP = form.watch("Protocol") === "tcp";
            const useTLS = form.watch("UseTLS");
            const tlsVerify = form.watch("TLSVerify");
            return isTCP && useTLS && tlsVerify;
          }}
        />

        {/* Client Certificate for mTLS */}
        <CertificateField
          form={form}
          name="ClientCertPath"
          label="Client Certificate (mTLS - Optional)"
          labelIcon={<ShieldCheckIcon className="w-3.5 h-3.5" />}
          icon={<FileIcon className="w-4 h-4 text-blue-500 flex-shrink-0" />}
          emptyText="No client certificate"
          onSelect={handleSelectClientCert}
          onClear={handleClearClientCert}
          showWhen={() => {
            const isTCP = form.watch("Protocol") === "tcp";
            const useTLS = form.watch("UseTLS");
            return isTCP && useTLS;
          }}
        />

        {/* Client Key for mTLS */}
        <CertificateField
          form={form}
          name="ClientKeyPath"
          label="Client Private Key (mTLS - Optional)"
          labelIcon={<KeyIcon className="w-3.5 h-3.5" />}
          icon={<KeyIcon className="w-4 h-4 text-amber-500 flex-shrink-0" />}
          emptyText="No client key"
          description="Both cert and key required for mutual TLS authentication"
          onSelect={handleSelectClientKey}
          onClear={handleClearClientKey}
          showWhen={() => {
            const isTCP = form.watch("Protocol") === "tcp";
            const useTLS = form.watch("UseTLS");
            return isTCP && useTLS;
          }}
        />
      </CardContent>
    </Card>
  );
}

// ================================================================================
// CERTIFICATE FIELD SUBCOMPONENT
// Single Responsibility: Renders a certificate selector field
// ================================================================================

interface CertificateFieldProps {
  form: UseFormReturn<FormData>;
  name: "CACertPath" | "ClientCertPath" | "ClientKeyPath";
  label: string;
  labelIcon?: React.ReactNode;
  icon: React.ReactNode;
  emptyText: string;
  description?: string;
  onSelect: () => Promise<void>;
  onClear: () => void;
  showWhen: () => boolean;
}

function CertificateField({
  form,
  name,
  label,
  labelIcon,
  icon,
  emptyText,
  description,
  onSelect,
  onClear,
  showWhen,
}: CertificateFieldProps) {
  return (
    <FormField
      control={form.control}
      name={name}
      render={({ field }) => {
        if (!showWhen()) return <></>;

        return (
          <FormItem>
            <FormLabel className="text-xs font-medium flex items-center gap-1.5">
              {labelIcon}
              {label}
            </FormLabel>
            <div className="flex gap-2 items-center">
              <FormControl>
                <div className="flex-1 flex items-center gap-2 h-9 px-3 border border-border/60 rounded-md bg-secondary/20">
                  {field.value ? (
                    <>
                      {icon}
                      <span className="truncate flex-1 text-sm">
                        {field.value.split("/").pop() || field.value.split("\\").pop()}
                      </span>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 p-0 hover:bg-destructive/10"
                        onClick={onClear}
                      >
                        <XIcon className="w-3.5 h-3.5 text-destructive" />
                      </Button>
                    </>
                  ) : (
                    <span className="text-muted-foreground text-sm">{emptyText}</span>
                  )}
                </div>
              </FormControl>
              <Button type="button" variant="outline" className="h-9 px-4" onClick={onSelect}>
                <FolderOpenIcon className="w-4 h-4 mr-2" />
                Browse
              </Button>
            </div>
            {description && (
              <FormDescription className="text-[10px]">{description}</FormDescription>
            )}
            <FormMessage className="text-xs" />
          </FormItem>
        );
      }}
    />
  );
}
