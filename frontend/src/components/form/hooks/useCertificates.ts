"use client";

import { useCallback } from "react";
import { UseFormReturn } from "react-hook-form";
import {
  SelectCACertificate,
  SelectClientCertificate,
  SelectClientKey,
} from "@/wailsjs/go/main/ConnectionService";
import { toast } from "@/hooks/use-toast";
import { FormData } from "../types";

// ================================================================================
// USE CERTIFICATES HOOK
// Single Responsibility: Manages TLS certificate selection and clearing
// ================================================================================

export interface UseCertificatesReturn {
  // CA Certificate
  handleSelectCACert: () => Promise<void>;
  handleClearCACert: () => void;
  // Client Certificate (mTLS)
  handleSelectClientCert: () => Promise<void>;
  handleClearClientCert: () => void;
  // Client Key (mTLS)
  handleSelectClientKey: () => Promise<void>;
  handleClearClientKey: () => void;
}

export function useCertificates(form: UseFormReturn<FormData>): UseCertificatesReturn {
  // CA Certificate handlers
  const handleSelectCACert = useCallback(async () => {
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
  }, [form]);

  const handleClearCACert = useCallback(() => {
    form.setValue("CACertPath", "");
  }, [form]);

  // Client Certificate handlers
  const handleSelectClientCert = useCallback(async () => {
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
  }, [form]);

  const handleClearClientCert = useCallback(() => {
    form.setValue("ClientCertPath", "");
  }, [form]);

  // Client Key handlers
  const handleSelectClientKey = useCallback(async () => {
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
  }, [form]);

  const handleClearClientKey = useCallback(() => {
    form.setValue("ClientKeyPath", "");
  }, [form]);

  return {
    handleSelectCACert,
    handleClearCACert,
    handleSelectClientCert,
    handleClearClientCert,
    handleSelectClientKey,
    handleClearClientKey,
  };
}
