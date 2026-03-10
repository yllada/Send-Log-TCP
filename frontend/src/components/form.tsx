"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { toast } from "@/hooks/use-toast";
import { Button } from "@/components/ui/button";

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./ui/select";
import { Textarea } from "./ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

import { useState } from "react";
import {
  CheckConnection,
  SendSyslogMessages,
  Disconnect,
  SelectCACertificate,
  SelectClientCertificate,
  SelectClientKey,
} from "@/wailsjs/go/main/App";
import {
  CheckCircleIcon,
  XCircleIcon,
  SendIcon,
  NetworkIcon,
  FileIcon,
  FolderOpenIcon,
  XIcon,
  KeyIcon,
  ShieldCheckIcon,
  FileTextIcon,
} from "lucide-react";
import { ProfileManager } from "@/components/profile-manager";
import { TemplateManager } from "@/components/template-manager";
import { BatchImport } from "@/components/batch-import";

// IP validation regex (IPv4 and IPv6)
const ipRegex = /^(?:(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}|(?:[a-fA-F0-9]{1,4}:){7}[a-fA-F0-9]{1,4}|(?:[a-fA-F0-9]{1,4}:){1,7}:|(?:[a-fA-F0-9]{1,4}:){1,6}:[a-fA-F0-9]{1,4}|::(?:[a-fA-F0-9]{1,4}:){0,5}[a-fA-F0-9]{1,4}|[a-fA-F0-9]{1,4}::(?:[a-fA-F0-9]{1,4}:){0,4}[a-fA-F0-9]{1,4}|localhost)$/;

const FormSchema = z.object({
  Address: z.string().regex(ipRegex, { message: "Invalid IP address (IPv4 or IPv6)" }),
  Port: z.string({ message: "Port is required" })
    .refine((val: string) => {
      const num = parseInt(val, 10);
      return !isNaN(num) && num >= 1 && num <= 65535;
    }, { message: "Port must be between 1-65535" }),
  Protocol: z.string({ message: "Please select a protocol" }),
  Messages: z.string().min(1, { message: "At least one message is required" }),
  FramingMethod: z.enum(["octet-counting", "non-transparent"], {
    message: "Please select a framing method",
  }),
  Facility: z.number().min(0).max(23, "Facility must be between 0-23"),
  Severity: z.number().min(0).max(7, "Severity must be between 0-7"),
  Hostname: z.string().optional(),
  Appname: z.string().min(1, "Application name is required"),
  UseRFC5424: z.boolean(),
  UseTLS: z.boolean(),
  TLSVerify: z.boolean(),
  CACertPath: z.string().optional(),
  ClientCertPath: z.string().optional(),
  ClientKeyPath: z.string().optional(),
});

export function InputForm() {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      Address: "",
      Port: "",
      Protocol: "",
      Messages: "",
      FramingMethod: "octet-counting",
      Facility: 16, // local0
      Severity: 6, // info
      Hostname: "",
      Appname: "sendlog",
      UseRFC5424: true,
      UseTLS: false,
      TLSVerify: false,
      CACertPath: "",
      ClientCertPath: "",
      ClientKeyPath: "",
    },
  });

  const [isConnected, setIsConnected] = useState(false);

  // Função para selecionar certificado CA
  const handleSelectCACert = async () => {
    try {
      const path = await SelectCACertificate();
      if (path) {
        form.setValue("CACertPath", path);
        toast({
          title: "✓ Certificate Selected",
          description: `CA certificate loaded: ${path.split('/').pop() || path.split('\\').pop()}`,
          variant: "default",
        });
      }
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to select certificate",
        variant: "destructive",
      });
    }
  };

  // Função para limpar certificado CA
  const handleClearCACert = () => {
    form.setValue("CACertPath", "");
  };

  // Función para seleccionar certificado del cliente (mTLS)
  const handleSelectClientCert = async () => {
    try {
      const path = await SelectClientCertificate();
      if (path) {
        form.setValue("ClientCertPath", path);
        toast({
          title: "✓ Client Certificate Selected",
          description: `Client cert loaded: ${path.split('/').pop() || path.split('\\').pop()}`,
          variant: "default",
        });
      }
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to select client certificate",
        variant: "destructive",
      });
    }
  };

  // Función para limpiar certificado del cliente
  const handleClearClientCert = () => {
    form.setValue("ClientCertPath", "");
  };

  // Función para seleccionar clave privada del cliente (mTLS)
  const handleSelectClientKey = async () => {
    try {
      const path = await SelectClientKey();
      if (path) {
        form.setValue("ClientKeyPath", path);
        toast({
          title: "✓ Client Key Selected",
          description: `Client key loaded: ${path.split('/').pop() || path.split('\\').pop()}`,
          variant: "default",
        });
      }
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to select client key",
        variant: "destructive",
      });
    }
  };

  // Función para limpiar clave del cliente
  const handleClearClientKey = () => {
    form.setValue("ClientKeyPath", "");
  };

  const handleConnectionToggle = async () => {
    const { Address, Port, Protocol, UseTLS, TLSVerify, CACertPath, ClientCertPath, ClientKeyPath } = form.getValues();

    if (isConnected) {
      await Disconnect();
      setIsConnected(false);
      toast({
        title: "✓ Disconnected",
        description: `Closed connection to ${Address}:${Port}`,
        variant: "default",
      });
    } else {
      try {
        const response = await CheckConnection(
          Address, 
          Port, 
          Protocol, 
          UseTLS, 
          TLSVerify, 
          CACertPath || "", 
          ClientCertPath || "", 
          ClientKeyPath || ""
        );
        if (response) {
          setIsConnected(true);
          const protocolInfo = UseTLS ? `${Protocol.toUpperCase()}+TLS` : Protocol.toUpperCase();
          const securityInfo = UseTLS && !TLSVerify ? " (self-signed cert)" : "";
          const caInfo = UseTLS && TLSVerify && CACertPath ? " (custom CA)" : "";
          const mTLSInfo = UseTLS && ClientCertPath && ClientKeyPath ? " (mTLS)" : "";
          toast({
            title: "✓ Connected",
            description: `Successfully connected to ${Address}:${Port} via ${protocolInfo}${securityInfo}${caInfo}${mTLSInfo}`,
            variant: "default",
          });
        } else {
          setIsConnected(false);
          toast({
            title: "✗ Connection Failed",
            description: "Unable to establish connection",
            variant: "destructive",
          });
        }
      } catch (error) {
        setIsConnected(false);
        toast({
          title: "✗ Connection Error",
          description: error instanceof Error ? error.message : "Unknown error",
          variant: "destructive",
        });
      }
    }
  };

  async function onSubmit(data: z.infer<typeof FormSchema>) {
    const updatedData = {
      Address: data.Address,
      Port: data.Port,
      Protocol: data.Protocol,
      Messages: data.Messages.split("\n").filter((msg: string) => msg.trim() !== ""),
      FramingMethod: data.FramingMethod,
      Facility: data.Facility,
      Severity: data.Severity,
      Hostname: data.Hostname || "",
      Appname: data.Appname,
      UseRFC5424: data.UseRFC5424,
      UseTLS: data.UseTLS,
      TLSVerify: data.TLSVerify,
      CACertPath: data.CACertPath || "",
      ClientCertPath: data.ClientCertPath || "",
      ClientKeyPath: data.ClientKeyPath || "",
    };

    try {
      const response = await SendSyslogMessages(updatedData);

      const successCount = response.sentMessages?.length || 0;
      const errorCount = response.errors?.length || 0;

      if (errorCount === 0 && successCount > 0) {
        toast({
          title: "✓ Messages Sent",
          description: `Successfully sent ${successCount} message${successCount > 1 ? 's' : ''}`,
          variant: "default",
        });
      } else if (errorCount > 0 && successCount > 0) {
        toast({
          title: "⚠ Partial Success",
          description: `Sent: ${successCount} | Failed: ${errorCount}`,
          variant: "default",
        });
      } else if (errorCount > 0) {
        toast({
          title: "✗ Send Failed",
          description: response.errors?.[0] || "Failed to send messages",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to send messages",
        variant: "destructive",
      });
    }
  }

  // Facility labels según RFC 5424
  const facilityOptions = [
    { value: 0, label: "0 - kernel" },
    { value: 1, label: "1 - user" },
    { value: 2, label: "2 - mail" },
    { value: 3, label: "3 - system daemon" },
    { value: 4, label: "4 - security/auth" },
    { value: 5, label: "5 - syslogd" },
    { value: 6, label: "6 - line printer" },
    { value: 7, label: "7 - network news" },
    { value: 8, label: "8 - UUCP" },
    { value: 9, label: "9 - clock daemon" },
    { value: 10, label: "10 - security/auth" },
    { value: 11, label: "11 - FTP daemon" },
    { value: 12, label: "12 - NTP" },
    { value: 13, label: "13 - log audit" },
    { value: 14, label: "14 - log alert" },
    { value: 15, label: "15 - clock daemon" },
    { value: 16, label: "16 - local0" },
    { value: 17, label: "17 - local1" },
    { value: 18, label: "18 - local2" },
    { value: 19, label: "19 - local3" },
    { value: 20, label: "20 - local4" },
    { value: 21, label: "21 - local5" },
    { value: 22, label: "22 - local6" },
    { value: 23, label: "23 - local7" },
  ];

  // Severity labels según RFC 5424
  const severityOptions = [
    { value: 0, label: "0 - Emergency" },
    { value: 1, label: "1 - Alert" },
    { value: 2, label: "2 - Critical" },
    { value: 3, label: "3 - Error" },
    { value: 4, label: "4 - Warning" },
    { value: 5, label: "5 - Notice" },
    { value: 6, label: "6 - Informational" },
    { value: 7, label: "7 - Debug" },
  ];

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="w-full max-w-3xl mx-auto space-y-3"
      >
        {/* Connection Settings Card */}
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
                      <Input 
                        placeholder="192.168.1.100" 
                        className="h-9" 
                        {...field} 
                      />
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
                      <Input 
                        placeholder="514" 
                        type="number" 
                        className="h-9"
                        {...field} 
                      />
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
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
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
                  onClick={handleConnectionToggle}
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
                    <FormLabel className={`text-xs font-medium ${!isTCP ? 'text-muted-foreground' : ''}`}>
                      TCP Framing
                    </FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      disabled={!isTCP}
                    >
                      <FormControl>
                        <SelectTrigger 
                          className="h-9"
                          disabled={!isTCP}
                        >
                          <SelectValue placeholder="Select method" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="octet-counting">
                          Octet Counting (RFC 6587)
                        </SelectItem>
                        <SelectItem value="non-transparent">
                          Non-Transparent (LF)
                        </SelectItem>
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
                            if (e.target.checked && form.getValues("Port") === "514") {
                              form.setValue("Port", "6514");
                            } else if (!e.target.checked && form.getValues("Port") === "6514") {
                              form.setValue("Port", "514");
                            }
                          }}
                          disabled={!isTCP}
                          className="h-4 w-4 rounded accent-primary"
                        />
                      </FormControl>
                      <div className="flex-1 space-y-0.5">
                        <FormLabel className={`text-xs font-medium leading-none ${!isTCP ? 'text-muted-foreground' : ''}`}>
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
                        <FormLabel className={`text-xs font-medium leading-none ${!isTCP || !useTLS ? 'text-muted-foreground' : ''}`}>
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

            {/* CA Certificate Selector - Only visible when TLS + Verify is enabled */}
            <FormField
              control={form.control}
              name="CACertPath"
              render={({ field }) => {
                const useTLS = form.watch("UseTLS");
                const tlsVerify = form.watch("TLSVerify");
                const isTCP = form.watch("Protocol") === "tcp";
                const showSelector = isTCP && useTLS && tlsVerify;
                
                if (!showSelector) return <></>;
                
                return (
                  <FormItem>
                    <FormLabel className="text-xs font-medium">CA Certificate (Optional)</FormLabel>
                    <div className="flex gap-2 items-center">
                      <FormControl>
                        <div className="flex-1 flex items-center gap-2 h-9 px-3 border border-border/60 rounded-md bg-secondary/20">
                          {field.value ? (
                            <>
                              <FileIcon className="w-4 h-4 text-green-500 flex-shrink-0" />
                              <span className="truncate flex-1 text-sm">
                                {field.value.split('/').pop() || field.value.split('\\').pop()}
                              </span>
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6 p-0 hover:bg-destructive/10"
                                onClick={handleClearCACert}
                              >
                                <XIcon className="w-3.5 h-3.5 text-destructive" />
                              </Button>
                            </>
                          ) : (
                            <span className="text-muted-foreground text-sm">
                              Using system CA store
                            </span>
                          )}
                        </div>
                      </FormControl>
                      <Button
                        type="button"
                        variant="outline"
                        className="h-9 px-4"
                        onClick={handleSelectCACert}
                      >
                        <FolderOpenIcon className="w-4 h-4 mr-2" />
                        Browse
                      </Button>
                    </div>
                    <FormDescription className="text-[10px]">
                      Load a custom CA certificate (.pem, .crt) for server verification
                    </FormDescription>
                    <FormMessage className="text-xs" />
                  </FormItem>
                );
              }}
            />

            {/* Client Certificate for mTLS */}
            <FormField
              control={form.control}
              name="ClientCertPath"
              render={({ field }) => {
                const useTLS = form.watch("UseTLS");
                const isTCP = form.watch("Protocol") === "tcp";
                const showSelector = isTCP && useTLS;
                
                if (!showSelector) return <></>;
                
                return (
                  <FormItem>
                    <FormLabel className="text-xs font-medium flex items-center gap-1.5">
                      <ShieldCheckIcon className="w-3.5 h-3.5" />
                      Client Certificate (mTLS - Optional)
                    </FormLabel>
                    <div className="flex gap-2 items-center">
                      <FormControl>
                        <div className="flex-1 flex items-center gap-2 h-9 px-3 border border-border/60 rounded-md bg-secondary/20">
                          {field.value ? (
                            <>
                              <FileIcon className="w-4 h-4 text-blue-500 flex-shrink-0" />
                              <span className="truncate flex-1 text-sm">
                                {field.value.split('/').pop() || field.value.split('\\').pop()}
                              </span>
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6 p-0 hover:bg-destructive/10"
                                onClick={handleClearClientCert}
                              >
                                <XIcon className="w-3.5 h-3.5 text-destructive" />
                              </Button>
                            </>
                          ) : (
                            <span className="text-muted-foreground text-sm">
                              No client certificate
                            </span>
                          )}
                        </div>
                      </FormControl>
                      <Button
                        type="button"
                        variant="outline"
                        className="h-9 px-4"
                        onClick={handleSelectClientCert}
                      >
                        <FolderOpenIcon className="w-4 h-4 mr-2" />
                        Browse
                      </Button>
                    </div>
                    <FormMessage className="text-xs" />
                  </FormItem>
                );
              }}
            />

            {/* Client Key for mTLS */}
            <FormField
              control={form.control}
              name="ClientKeyPath"
              render={({ field }) => {
                const useTLS = form.watch("UseTLS");
                const isTCP = form.watch("Protocol") === "tcp";
                const showSelector = isTCP && useTLS;
                
                if (!showSelector) return <></>;
                
                return (
                  <FormItem>
                    <FormLabel className="text-xs font-medium flex items-center gap-1.5">
                      <KeyIcon className="w-3.5 h-3.5" />
                      Client Private Key (mTLS - Optional)
                    </FormLabel>
                    <div className="flex gap-2 items-center">
                      <FormControl>
                        <div className="flex-1 flex items-center gap-2 h-9 px-3 border border-border/60 rounded-md bg-secondary/20">
                          {field.value ? (
                            <>
                              <KeyIcon className="w-4 h-4 text-amber-500 flex-shrink-0" />
                              <span className="truncate flex-1 text-sm">
                                {field.value.split('/').pop() || field.value.split('\\').pop()}
                              </span>
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6 p-0 hover:bg-destructive/10"
                                onClick={handleClearClientKey}
                              >
                                <XIcon className="w-3.5 h-3.5 text-destructive" />
                              </Button>
                            </>
                          ) : (
                            <span className="text-muted-foreground text-sm">
                              No client key
                            </span>
                          )}
                        </div>
                      </FormControl>
                      <Button
                        type="button"
                        variant="outline"
                        className="h-9 px-4"
                        onClick={handleSelectClientKey}
                      >
                        <FolderOpenIcon className="w-4 h-4 mr-2" />
                        Browse
                      </Button>
                    </div>
                    <FormDescription className="text-[10px] text-muted-foreground">
                      Both cert and key required for mutual TLS authentication
                    </FormDescription>
                    <FormMessage className="text-xs" />
                  </FormItem>
                );
              }}
            />
          </CardContent>
        </Card>

        {/* Message Format Card */}
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
                        {facilityOptions.map((option) => (
                          <SelectItem
                            key={option.value}
                            value={option.value.toString()}
                          >
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
                        {severityOptions.map((option) => (
                          <SelectItem
                            key={option.value}
                            value={option.value.toString()}
                          >
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

            <div className="grid grid-cols-2 gap-3">
              <FormField
                control={form.control}
                name="Hostname"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="text-xs font-medium">Hostname</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Optional"
                        className="h-9"
                        {...field}
                      />
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
                      <Input 
                        placeholder="sendlog" 
                        className="h-9"
                        {...field} 
                      />
                    </FormControl>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Messages Card */}
        {isConnected ? (
          <Card>
            <CardHeader className="pb-2 pt-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-semibold flex items-center gap-2">
                  <SendIcon className="w-4 h-4" />
                  Send Messages
                </CardTitle>
                <BatchImport
                  currentMessages={form.watch("Messages")}
                  onImport={(messages) => {
                    form.setValue("Messages", messages.join("\n"));
                  }}
                />
              </div>
            </CardHeader>
            <CardContent className="space-y-3 pb-3">
              <FormField
                control={form.control}
                name="Messages"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="text-xs font-medium">Log Messages</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Type your messages here, one per line..."
                        className="resize-none h-[80px]"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription className="text-[10px] text-muted-foreground">
                      Enter messages separated by new lines
                    </FormDescription>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />

              <Button type="submit" className="w-full h-9">
                <SendIcon className="w-4 h-4 mr-2" />
                Send Syslog Messages
              </Button>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent className="py-6">
              <div className="text-center">
                <NetworkIcon className="w-10 h-10 mx-auto mb-2 text-muted-foreground/50" />
                <p className="text-sm text-muted-foreground">
                  Connect to a syslog server to send messages
                </p>
              </div>
            </CardContent>
          </Card>
        )}
      </form>
    </Form>
  );
}
