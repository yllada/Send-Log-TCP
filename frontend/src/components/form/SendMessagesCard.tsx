"use client";

import { UseFormReturn } from "react-hook-form";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { SendIcon, NetworkIcon } from "lucide-react";
import { BatchImport } from "@/components/batch-import";
import { FormData } from "./types";

// ================================================================================
// SEND MESSAGES CARD COMPONENT
// Single Responsibility: Handles the message input and send UI
// ================================================================================

interface SendMessagesCardProps {
  form: UseFormReturn<FormData>;
  isConnected: boolean;
}

export function SendMessagesCard({ form, isConnected }: SendMessagesCardProps) {
  if (!isConnected) {
    return <DisconnectedState />;
  }

  return (
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
  );
}

// ================================================================================
// DISCONNECTED STATE SUBCOMPONENT
// ================================================================================

function DisconnectedState() {
  return (
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
  );
}
