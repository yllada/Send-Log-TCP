"use client";

import { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "@/hooks/use-toast";
import {
  GetProfiles,
  SaveProfile,
  DeleteProfile,
} from "@/wailsjs/go/main/App";
import {
  FolderOpenIcon,
  SaveIcon,
  Trash2Icon,
  PlusIcon,
  ServerIcon,
} from "lucide-react";

// Profile type matching Go struct
interface ConnectionProfile {
  id: string;
  name: string;
  description?: string;
  address: string;
  port: string;
  protocol: string;
  framingMethod: string;
  useTls: boolean;
  tlsVerify: boolean;
  caCertPath?: string;
  clientCertPath?: string;
  clientKeyPath?: string;
  createdAt: number;
  updatedAt: number;
}

interface ProfileManagerProps {
  // Current form values to save
  currentValues: {
    Address: string;
    Port: string;
    Protocol: string;
    FramingMethod: string;
    UseTLS: boolean;
    TLSVerify: boolean;
    CACertPath?: string;
    ClientCertPath?: string;
    ClientKeyPath?: string;
  };
  // Callback to load profile into form
  onLoadProfile: (profile: ConnectionProfile) => void;
}

export function ProfileManager({ currentValues, onLoadProfile }: ProfileManagerProps) {
  const [profiles, setProfiles] = useState<ConnectionProfile[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isSaveOpen, setIsSaveOpen] = useState(false);
  const [profileName, setProfileName] = useState("");
  const [profileDescription, setProfileDescription] = useState("");
  const [selectedProfileId, setSelectedProfileId] = useState<string>("");

  // Load profiles on mount
  const loadProfiles = useCallback(async () => {
    try {
      const data = await GetProfiles();
      setProfiles(data || []);
    } catch (error) {
      console.error("Failed to load profiles:", error);
    }
  }, []);

  useEffect(() => {
    loadProfiles();
  }, [loadProfiles]);

  // Save current connection as profile
  const handleSave = async () => {
    if (!profileName.trim()) {
      toast({
        title: "✗ Error",
        description: "Profile name is required",
        variant: "destructive",
      });
      return;
    }

    try {
      const profile: Partial<ConnectionProfile> = {
        name: profileName.trim(),
        description: profileDescription.trim(),
        address: currentValues.Address,
        port: currentValues.Port,
        protocol: currentValues.Protocol,
        framingMethod: currentValues.FramingMethod,
        useTls: currentValues.UseTLS,
        tlsVerify: currentValues.TLSVerify,
        caCertPath: currentValues.CACertPath,
        clientCertPath: currentValues.ClientCertPath,
        clientKeyPath: currentValues.ClientKeyPath,
      };

      await SaveProfile(profile as ConnectionProfile);
      await loadProfiles();

      toast({
        title: "✓ Profile Saved",
        description: `"${profileName}" saved successfully`,
        variant: "default",
      });

      setProfileName("");
      setProfileDescription("");
      setIsSaveOpen(false);
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to save profile",
        variant: "destructive",
      });
    }
  };

  // Load selected profile
  const handleLoad = () => {
    const profile = profiles.find((p) => p.id === selectedProfileId);
    if (profile) {
      onLoadProfile(profile);
      toast({
        title: "✓ Profile Loaded",
        description: `"${profile.name}" loaded`,
        variant: "default",
      });
      setIsOpen(false);
    }
  };

  // Delete profile
  const handleDelete = async () => {
    if (!selectedProfileId) return;

    const profile = profiles.find((p) => p.id === selectedProfileId);
    if (!profile) return;

    try {
      await DeleteProfile(selectedProfileId);
      await loadProfiles();
      setSelectedProfileId("");

      toast({
        title: "✓ Profile Deleted",
        description: `"${profile.name}" deleted`,
        variant: "default",
      });
    } catch (error) {
      toast({
        title: "✗ Error",
        description: error instanceof Error ? error.message : "Failed to delete profile",
        variant: "destructive",
      });
    }
  };

  return (
    <div className="flex items-center gap-1">
      {/* Load Profile Button */}
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger asChild>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="h-9 w-9"
            title="Load Profile"
          >
            <FolderOpenIcon className="w-4 h-4" />
          </Button>
        </DialogTrigger>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <ServerIcon className="w-5 h-5" />
              Connection Profiles
            </DialogTitle>
            <DialogDescription>
              Load a saved connection configuration
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {profiles.length === 0 ? (
              <div className="text-center text-muted-foreground py-8">
                No saved profiles
              </div>
            ) : (
              <div className="space-y-2">
                <Label>Select Profile</Label>
                <Select
                  value={selectedProfileId}
                  onValueChange={setSelectedProfileId}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Choose a profile..." />
                  </SelectTrigger>
                  <SelectContent>
                    {profiles.map((profile) => (
                      <SelectItem key={profile.id} value={profile.id}>
                        <div className="flex flex-col">
                          <span>{profile.name}</span>
                          <span className="text-xs text-muted-foreground">
                            {profile.address}:{profile.port} ({profile.protocol.toUpperCase()})
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                {selectedProfileId && (
                  <div className="mt-4 p-3 bg-muted rounded-lg text-sm space-y-1">
                    {(() => {
                      const p = profiles.find((p) => p.id === selectedProfileId);
                      if (!p) return null;
                      return (
                        <>
                          <div><strong>Server:</strong> {p.address}:{p.port}</div>
                          <div><strong>Protocol:</strong> {p.protocol.toUpperCase()}</div>
                          <div><strong>TLS:</strong> {p.useTls ? "Yes" : "No"}</div>
                          {p.description && (
                            <div><strong>Description:</strong> {p.description}</div>
                          )}
                        </>
                      );
                    })()}
                  </div>
                )}
              </div>
            )}
          </div>

          <DialogFooter className="gap-2 sm:gap-0">
            {selectedProfileId && (
              <Button
                type="button"
                variant="destructive"
                size="sm"
                onClick={handleDelete}
              >
                <Trash2Icon className="w-4 h-4 mr-1" />
                Delete
              </Button>
            )}
            <Button
              type="button"
              onClick={handleLoad}
              disabled={!selectedProfileId}
            >
              <FolderOpenIcon className="w-4 h-4 mr-1" />
              Load
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Save Profile Button */}
      <Dialog open={isSaveOpen} onOpenChange={setIsSaveOpen}>
        <DialogTrigger asChild>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="h-9 w-9"
            title="Save Profile"
          >
            <SaveIcon className="w-4 h-4" />
          </Button>
        </DialogTrigger>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <PlusIcon className="w-5 h-5" />
              Save Connection Profile
            </DialogTitle>
            <DialogDescription>
              Save current connection settings for later use
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="profileName">Profile Name *</Label>
              <Input
                id="profileName"
                value={profileName}
                onChange={(e) => setProfileName(e.target.value)}
                placeholder="e.g., Production Server"
                className="h-9"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profileDesc">Description (optional)</Label>
              <Input
                id="profileDesc"
                value={profileDescription}
                onChange={(e) => setProfileDescription(e.target.value)}
                placeholder="e.g., Main syslog server"
                className="h-9"
              />
            </div>

            <div className="p-3 bg-muted rounded-lg text-sm space-y-1">
              <div className="font-medium mb-2">Connection Details:</div>
              <div><strong>Server:</strong> {currentValues.Address || "(not set)"}:{currentValues.Port || "(not set)"}</div>
              <div><strong>Protocol:</strong> {currentValues.Protocol?.toUpperCase() || "(not set)"}</div>
              <div><strong>TLS:</strong> {currentValues.UseTLS ? "Yes" : "No"}</div>
            </div>
          </div>

          <DialogFooter>
            <Button type="button" onClick={handleSave}>
              <SaveIcon className="w-4 h-4 mr-1" />
              Save Profile
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
