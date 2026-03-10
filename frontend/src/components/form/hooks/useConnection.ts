"use client";

import { useState, useCallback } from "react";
import { UseFormReturn } from "react-hook-form";
import {
  CheckConnection,
  Disconnect,
} from "@/wailsjs/go/main/ConnectionService";
import { toast } from "@/hooks/use-toast";
import { FormData } from "../types";

// ================================================================================
// USE CONNECTION HOOK
// Single Responsibility: Manages connection state and operations
// ================================================================================

export interface UseConnectionReturn {
  isConnected: boolean;
  handleConnectionToggle: () => Promise<void>;
}

export function useConnection(form: UseFormReturn<FormData>): UseConnectionReturn {
  const [isConnected, setIsConnected] = useState(false);

  const handleConnectionToggle = useCallback(async () => {
    const { 
      Address, 
      Port, 
      Protocol, 
      UseTLS, 
      TLSVerify, 
      CACertPath, 
      ClientCertPath, 
      ClientKeyPath 
    } = form.getValues();

    if (isConnected) {
      // Disconnect
      await Disconnect();
      setIsConnected(false);
      toast({
        title: "✓ Disconnected",
        description: `Closed connection to ${Address}:${Port}`,
        variant: "default",
      });
    } else {
      // Connect
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
  }, [form, isConnected]);

  return {
    isConnected,
    handleConnectionToggle,
  };
}
