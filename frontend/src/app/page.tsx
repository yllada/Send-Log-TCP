"use client";

import { InputForm } from "@/components/form";
import { StressTestForm } from "@/components/stress-test";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SendIcon, Timer } from "lucide-react";

export default function Home() {
  return (
    <div className="max-w-3xl w-full">
      <Tabs defaultValue="send" className="w-full">
        <TabsList className="w-full grid grid-cols-2 mb-3">
          <TabsTrigger value="send" className="flex items-center gap-2">
            <SendIcon className="w-4 h-4" />
            Send Messages
          </TabsTrigger>
          <TabsTrigger value="stress" className="flex items-center gap-2">
            <Timer className="w-4 h-4" />
            Stress Test
          </TabsTrigger>
        </TabsList>
        <TabsContent value="send">
          <InputForm />
        </TabsContent>
        <TabsContent value="stress">
          <StressTestForm />
        </TabsContent>
      </Tabs>
    </div>
  );
}
