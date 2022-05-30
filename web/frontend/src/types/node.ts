export const enum NodeStatus {
  // Internal Status when the actor hasn't been initialized yet
  NotInitialized = -1,
  // Initialized is when the actor has been initialized
  Initialized,
  // Setup is when the actor was set up
  Setup,
  // Failed is when the actor failed to set up
  Failed,
}

interface DKGInfo {
  Status: NodeStatus.Initialized | NodeStatus.Setup | NodeStatus.Failed;
  Error: { Title: string; Code: number; Message: string; Args: string[] };
}

interface NodeProxyAddress {
  NodeAddr: string;
  Proxy: string;
}

export type { DKGInfo, NodeProxyAddress };
