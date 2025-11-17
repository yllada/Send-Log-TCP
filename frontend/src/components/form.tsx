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

import { useState } from "react";
import {
  CheckConnection,
  SendSyslogMessages,
  Disconnect,
} from "@/wailsjs/go/main/App";
import {
  CheckIcon,
  ChevronRightIcon,
  CrossCircledIcon,
} from "@radix-ui/react-icons";
import { CheckCircle, CheckCircleIcon, XIcon } from "lucide-react";

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
    },
  });

  const [isConnected, setIsConnected] = useState(false);

  const handleConnectionToggle = async () => {
    const { Address, Port, Protocol } = form.getValues();

    if (isConnected) {
      // Lógica para desconectar
      // Por ejemplo, podrías llamar a un método para cerrar la conexión en el backend
      await Disconnect(); // Asegúrate de que este método esté definido
      setIsConnected(false);
      toast({
        title: "Disconnected successfully!",
        description: (
          <p>
            {" "}
            Disconnected to server:{" "}
            <span className="text-red-500">{Address}</span> on port:{" "}
            <span className="text-red-500">{Port}</span>
          </p>
        ),
      });
    } else {
      // Lógica para conectar
      try {
        const response = await CheckConnection(Address, Port, Protocol);
        if (response) {
          setIsConnected(true);
          toast({
            title: "Connected successfully!",
            description:(
              <p>
            {" "}
            Connected to server:{" "}
            <span className="text-red-500">{Address}</span> on port:{" "}
            <span className="text-red-500">{Port}</span>
          </p>
            ),
          });
        } else {
          setIsConnected(false);
          toast({ title: "Failed to connect!" });
        }
      } catch (error) {
        setIsConnected(false);
        toast({ title: "Error connecting!" });
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
    };

    try {
      console.log("Attempting to send log...");

      // Llamada al backend para obtener los mensajes enviados y errores
      const response = await SendSyslogMessages(updatedData);
      console.log("Log sent successfully: ", response);

      // Mostrar el resultado en el toast
      toast({
        title: "Syslog Response:",
        description: (
          <pre className="mt-2 w-[340px] rounded-md bg-slate-950 p-4">
            <code className="text-white">
              Sent Messages: {JSON.stringify(response.sentMessages, null, 2)}
              {"\n"}
              Errors: {JSON.stringify(response.errors, null, 2)}
            </code>
          </pre>
        ),
      });
    } catch (error) {
      console.error("Error sending log: ", error);

      // Mostrar un toast en caso de error
      toast({
        title: "Error",
        description: "There was an issue sending the log. Please try again.",
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
        className="space-y-6 p-6 bg-gray-100 rounded-lg shadow-md"
      >
        {/* Connection Settings Section */}
        <div className="space-y-4">
          <h2 className="text-xl font-semibold text-gray-700">
            Connection Settings
          </h2>

          <div className="flex space-x-4">
            <FormField
              control={form.control}
              name="Address"
              render={({ field }) => (
                <FormItem className="flex-1">
                  <FormLabel>Address</FormLabel>
                  <FormControl>
                    <Input placeholder="192.168.1.100" {...field} />
                  </FormControl>
                  <FormDescription>
                    Enter the Syslog server IP address.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="Port"
              render={({ field }) => (
                <FormItem className="flex-1">
                  <FormLabel>Port</FormLabel>
                  <FormControl>
                    <Input placeholder="514" type="number" {...field} />
                  </FormControl>
                  <FormDescription>Enter the Syslog server port.</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className="flex items-center space-x-4">
            <FormField
              control={form.control}
              name="Protocol"
              render={({ field }) => (
                <FormItem className="flex-1">
                  <FormLabel>Protocol</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a protocol" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="tcp">TCP</SelectItem>
                      <SelectItem value="udp">UDP</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Select the protocol for Syslog communication.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button
              type="button"
              onClick={handleConnectionToggle}
              className={`rounded-full text-white shadow-md transition duration-200 mt-8 ${
                isConnected ? "bg-red-500 hover:bg-red-600" : "bg-green-500 hover:bg-green-600"
              }`}
            >
              {isConnected ? <CrossCircledIcon /> : <CheckCircleIcon />}
            </Button>
          </div>

          {/* TCP Framing Method - Only shown when protocol is TCP */}
          {form.watch("Protocol") === "tcp" && (
            <FormField
              control={form.control}
              name="FramingMethod"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Framing Method (TCP)</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select framing method" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="octet-counting">
                        Octet Counting (RFC 6587)
                      </SelectItem>
                      <SelectItem value="non-transparent">
                        Non-Transparent Framing (LF delimiter)
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    Select the TCP framing method. Octet counting is recommended
                    for binary data.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          )}
        </div>

        <Separator />

        {/* Syslog Message Format Section */}
        <div className="space-y-4">
          <h2 className="text-xl font-semibold text-gray-700">
            Message Format
          </h2>

          <div className="flex items-center space-x-4">
            <FormField
              control={form.control}
              name="UseRFC5424"
              render={({ field }) => (
                <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4">
                  <FormControl>
                    <input
                      type="checkbox"
                      checked={field.value}
                      onChange={field.onChange}
                      className="h-4 w-4 mt-1"
                    />
                  </FormControl>
                  <div className="space-y-1 leading-none">
                    <FormLabel>
                      Use RFC 5424 (Modern Format)
                    </FormLabel>
                    <FormDescription>
                      When checked, uses RFC 5424 format. Unchecked uses legacy RFC 3164.
                    </FormDescription>
                  </div>
                </FormItem>
              )}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <FormField
              control={form.control}
              name="Facility"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Facility</FormLabel>
                  <Select
                    onValueChange={(value) => field.onChange(parseInt(value))}
                    defaultValue={field.value?.toString()}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select facility" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
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
                  <FormDescription>
                    Select the syslog facility (0-23).
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="Severity"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Severity</FormLabel>
                  <Select
                    onValueChange={(value) => field.onChange(parseInt(value))}
                    defaultValue={field.value?.toString()}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select severity" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
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
                  <FormDescription>
                    Select the message severity (0-7).
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <FormField
              control={form.control}
              name="Hostname"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Hostname (Optional)</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="Leave empty to use system hostname"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Custom hostname for syslog messages.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="Appname"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Application Name</FormLabel>
                  <FormControl>
                    <Input placeholder="sendlog" {...field} />
                  </FormControl>
                  <FormDescription>
                    Name of the application sending logs.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>

        <Separator />

        {/* Messages Section */}
        {isConnected && (
          <div className="space-y-4">
            <h2 className="text-xl font-semibold text-gray-700">
              Messages
            </h2>

            <FormField
              control={form.control}
              name="Messages"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Syslog Messages</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Type your messages here, one per line."
                      className="min-h-[120px]"
                      {...field}
                      onChange={(e) => field.onChange([e.target.value])}
                    />
                  </FormControl>
                  <FormDescription>
                    Enter multiple messages, separated by new lines.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button
              type="submit"
              className="w-full bg-blue-600 hover:bg-blue-700 text-white"
            >
              Send Syslog Messages
            </Button>
          </div>
        )}

        {!isConnected && (
          <div className="text-center text-gray-500 py-4">
            Please connect to a Syslog server to send messages.
          </div>
        )}
      </form>
    </Form>
  );
}
