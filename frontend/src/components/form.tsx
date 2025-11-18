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
import { Separator } from "./ui/separator";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

import { useState } from "react";
import {
  CheckConnection,
  SendSyslogMessages,
  Disconnect,
} from "@/wailsjs/go/main/App";
import {
  CheckCircleIcon,
  XCircleIcon,
  SendIcon,
  NetworkIcon,
} from "lucide-react";

const FormSchema = z.object({
  Address: z.string().ip({ version: "v4", message: "Invalid IP address" }),
  Port: z.string({ message: "Port is required" }),
  Protocol: z.string({ message: "Please select a protocol" }),
  Messages: z.array(z.string({ message: "Messages cannot be empty" })),
  FramingMethod: z.enum(["octet-counting", "non-transparent"], {
    message: "Please select a framing method",
  }),
  Facility: z.coerce.number().min(0).max(23, "Facility must be between 0-23"),
  Severity: z.coerce.number().min(0).max(7, "Severity must be between 0-7"),
  Hostname: z.string().optional(),
  Appname: z.string().min(1, "Application name is required"),
  UseRFC5424: z.boolean(),
  UseTLS: z.boolean(),
  TLSVerify: z.boolean(),
});

export function InputForm() {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      Address: "",
      Port: "",
      Protocol: "",
      Messages: [],
      FramingMethod: "octet-counting",
      Facility: 16, // local0
      Severity: 6, // info
      Hostname: "",
      Appname: "sendlog",
      UseRFC5424: true,
      UseTLS: false,
      TLSVerify: false,
    },
  });

  const [isConnected, setIsConnected] = useState(false);

  const handleConnectionToggle = async () => {
    const { Address, Port, Protocol, UseTLS, TLSVerify } = form.getValues();

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
        const response = await CheckConnection(Address, Port, Protocol, UseTLS, TLSVerify);
        if (response) {
          setIsConnected(true);
          const protocolInfo = UseTLS ? `${Protocol.toUpperCase()}+TLS` : Protocol.toUpperCase();
          const securityInfo = UseTLS && !TLSVerify ? " (self-signed cert)" : "";
          toast({
            title: "✓ Connected",
            description: `Successfully connected to ${Address}:${Port} via ${protocolInfo}${securityInfo}`,
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
      Messages: data.Messages[0].split("\n").filter((msg) => msg.trim() !== ""),
      FramingMethod: data.FramingMethod,
      Facility: data.Facility,
      Severity: data.Severity,
      Hostname: data.Hostname || "",
      Appname: data.Appname,
      UseRFC5424: data.UseRFC5424,
      UseTLS: data.UseTLS,
      TLSVerify: data.TLSVerify,
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
        className="w-full max-w-4xl mx-auto space-y-2"
      >
        {/* Connection Settings Card */}
        <Card>
          <CardHeader className="pb-2 pt-3">
            <CardTitle className="text-base flex items-center gap-2">
              <NetworkIcon className="w-4 h-4" />
              Connection Settings
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 pb-3">
            <div className="grid grid-cols-[2fr_1fr_1fr_auto] gap-2 items-end">
              <FormField
                control={form.control}
                name="Address"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="text-xs">IP Address</FormLabel>
                    <FormControl>
                      <Input 
                        placeholder="192.168.1.100" 
                        className="h-8 text-sm" 
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
                  <FormItem>
                    <FormLabel className="text-xs">Port</FormLabel>
                    <FormControl>
                      <Input 
                        placeholder="514" 
                        type="number" 
                        className="h-8 text-sm"
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
                  <FormItem>
                    <FormLabel className="text-xs">Protocol</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger className="h-8 text-sm">
                          <SelectValue placeholder="Protocol" />
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

              <Button
                type="button"
                onClick={handleConnectionToggle}
                size="icon"
                className={`h-8 w-8 rounded-full ${
                  isConnected
                    ? "bg-red-500 hover:bg-red-600"
                    : "bg-green-500 hover:bg-green-600"
                }`}
              >
                {isConnected ? (
                  <XCircleIcon className="w-4 h-4" />
                ) : (
                  <CheckCircleIcon className="w-4 h-4" />
                )}
              </Button>
            </div>

            <FormField
              control={form.control}
              name="FramingMethod"
              render={({ field }) => {
                const isTCP = form.watch("Protocol") === "tcp";
                
                return (
                  <FormItem>
                    <FormLabel className={`text-xs ${!isTCP ? 'text-muted-foreground' : ''}`}>
                      TCP Framing
                    </FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      disabled={!isTCP}
                    >
                      <FormControl>
                        <SelectTrigger 
                          className="h-8 text-sm"
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

            <div className="grid grid-cols-2 gap-2">
              <FormField
                control={form.control}
                name="UseTLS"
                render={({ field }) => {
                  const isTCP = form.watch("Protocol") === "tcp";
                  
                  return (
                    <FormItem className="flex flex-row items-center gap-2 space-y-0 rounded-md border p-2">
                      <FormControl>
                        <input
                          type="checkbox"
                          checked={field.value}
                          onChange={(e) => {
                            field.onChange(e.target.checked);
                            // Auto-sugerir puerto 6514 cuando se activa TLS
                            if (e.target.checked && form.getValues("Port") === "514") {
                              form.setValue("Port", "6514");
                            } else if (!e.target.checked && form.getValues("Port") === "6514") {
                              form.setValue("Port", "514");
                            }
                          }}
                          disabled={!isTCP}
                          className="h-3.5 w-3.5"
                        />
                      </FormControl>
                      <div className="flex-1 space-y-0">
                        <FormLabel className={`text-xs font-medium ${!isTCP ? 'text-muted-foreground' : ''}`}>
                          Use TLS/SSL
                        </FormLabel>
                        <FormDescription className="text-[10px] leading-tight">
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
                    <FormItem className="flex flex-row items-center gap-2 space-y-0 rounded-md border p-2">
                      <FormControl>
                        <input
                          type="checkbox"
                          checked={field.value}
                          onChange={field.onChange}
                          disabled={!isTCP || !useTLS}
                          className="h-3.5 w-3.5"
                        />
                      </FormControl>
                      <div className="flex-1 space-y-0">
                        <FormLabel className={`text-xs font-medium ${!isTCP || !useTLS ? 'text-muted-foreground' : ''}`}>
                          Verify Certificate
                        </FormLabel>
                        <FormDescription className="text-[10px] leading-tight">
                          Uncheck for self-signed certs
                        </FormDescription>
                      </div>
                    </FormItem>
                  );
                }}
              />
            </div>
          </CardContent>
        </Card>

        {/* Message Format Card */}
        <Card>
          <CardHeader className="pb-2 pt-3">
            <CardTitle className="text-base">Message Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 pb-3">
            <FormField
              control={form.control}
              name="UseRFC5424"
              render={({ field }) => (
                <FormItem className="flex flex-row items-center gap-2 space-y-0 rounded-md border p-2">
                  <FormControl>
                    <input
                      type="checkbox"
                      checked={field.value}
                      onChange={field.onChange}
                      className="h-3.5 w-3.5"
                    />
                  </FormControl>
                  <div className="flex-1 space-y-0">
                    <FormLabel className="text-xs font-medium">
                      Use RFC 5424 Format
                    </FormLabel>
                    <FormDescription className="text-[10px] leading-tight">
                      Modern format (checked) or legacy RFC 3164
                    </FormDescription>
                  </div>
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-2">
              <FormField
                control={form.control}
                name="Facility"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="text-xs">Facility</FormLabel>
                    <Select
                      onValueChange={(value) => field.onChange(parseInt(value))}
                      defaultValue={field.value?.toString()}
                    >
                      <FormControl>
                        <SelectTrigger className="h-8 text-sm">
                          <SelectValue placeholder="Select" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent className="max-h-[180px]">
                        {facilityOptions.map((option) => (
                          <SelectItem
                            key={option.value}
                            value={option.value.toString()}
                            className="text-xs"
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
                    <FormLabel className="text-xs">Severity</FormLabel>
                    <Select
                      onValueChange={(value) => field.onChange(parseInt(value))}
                      defaultValue={field.value?.toString()}
                    >
                      <FormControl>
                        <SelectTrigger className="h-8 text-sm">
                          <SelectValue placeholder="Select" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent className="max-h-[180px]">
                        {severityOptions.map((option) => (
                          <SelectItem
                            key={option.value}
                            value={option.value.toString()}
                            className="text-xs"
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

            <div className="grid grid-cols-2 gap-2">
              <FormField
                control={form.control}
                name="Hostname"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="text-xs">Hostname</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Optional"
                        className="h-8 text-sm"
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
                    <FormLabel className="text-xs">App Name</FormLabel>
                    <FormControl>
                      <Input 
                        placeholder="sendlog" 
                        className="h-8 text-sm"
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
              <CardTitle className="text-base flex items-center gap-2">
                <SendIcon className="w-4 h-4" />
                Send Messages
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 pb-3">
              <FormField
                control={form.control}
                name="Messages"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="text-xs">Log Messages</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Type your messages here, one per line..."
                        className="resize-none h-[90px] text-sm"
                        {...field}
                        onChange={(e) => field.onChange([e.target.value])}
                      />
                    </FormControl>
                    <FormDescription className="text-[10px]">
                      Enter messages separated by new lines
                    </FormDescription>
                    <FormMessage className="text-xs" />
                  </FormItem>
                )}
              />

              <Button type="submit" className="w-full h-8 text-sm" size="sm">
                <SendIcon className="w-3.5 h-3.5 mr-2" />
                Send Syslog Messages
              </Button>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent className="py-6">
              <div className="text-center">
                <NetworkIcon className="w-10 h-10 mx-auto mb-2 text-muted-foreground opacity-50" />
                <p className="text-xs text-muted-foreground">
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
