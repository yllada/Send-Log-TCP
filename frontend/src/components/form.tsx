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
});

export function InputForm() {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      Address: "",
      Port: "",
      Protocol: "",
      Messages: [],
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
      ...data,
      Messages: data.Messages[0].split("\n"),
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
        // status: "error",
      });
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="space-y-6 p-6 bg-gray-100 rounded-lg shadow-md"
      >
        <h2 className="text-xl font-semibold text-gray-700">
          Connect to Syslog Server
        </h2>

        <div className="flex space-x-4">
          <FormField
            control={form.control}
            name="Address"
            render={({ field }) => (
              <FormItem className="flex-1">
                <FormLabel>Address</FormLabel>
                <FormControl>
                  <Input placeholder="IP Address" {...field} />
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
                  <Input placeholder="Port" type="number" {...field} />
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
            className={`rounded-full  text-white shadow-md transition duration-200 ${
              isConnected ? "bg-red-500" : "bg-green-500"
            } text-white`}
          >
            {isConnected ? <CrossCircledIcon /> : <CheckCircleIcon />}
          </Button>
        </div>

        {isConnected && (
          <FormField
            control={form.control}
            name="Messages"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Messages</FormLabel>
                <FormControl>
                  <Textarea
                    placeholder="Type your messages here, one per line."
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
        )}

        {isConnected && (
          <Button
            type="submit"
            className="w-full bg-blue-600 hover:bg-blue-700 text-white"
          >
            Submit
          </Button>
        )}
      </form>
    </Form>
  );
}
