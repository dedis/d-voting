export const enum NodeStatus {
  // Internal Status when the proxy and/or its node do not respond
  Unreachable = -2,
  // Internal Status when the actor hasn't been initialized yet
  NotInitialized = -1,
  // Initialized is when the actor has been initialized
  Initialized,
  // Setup is when the actor was set up
  Setup,
  // Failed is when the actor failed to set up
  Failed,
  // Dealing is when the actor is sending its deals
  Dealing,
  // Responding is when the actor is sending its responses on the received deals
  Responding,
  // Certifying is when the actor is validating its deals based on the responses
  Certifying,
  // Certified is when the actor has been certified
  Certified,
}

interface DKGInfo {
  Status: NodeStatus.Initialized | NodeStatus.Setup | NodeStatus.Failed;
  Error: { Title: string; Code: number; Message: string; Args: string[] };
}

interface NodeProxyAddress {
  NodeAddr: string;
  Proxy: string;
}

// InternalDKGInfo is used to internally provide the status of DKG on a node.
class InternalDKGInfo {
  static fromStatus(status: NodeStatus): InternalDKGInfo {
    return new InternalDKGInfo(status, undefined);
  }

  static fromInfo(info: DKGInfo): InternalDKGInfo {
    return new InternalDKGInfo(info.Status, info.Error);
  }

  private constructor(private status: NodeStatus, private error: DKGInfo['Error']) {}

  getError(): string {
    if (this.error === undefined || this.error.Title === '') {
      return '';
    }

    return this.error.Title + ' - ' + this.error.Code + ' - ' + this.error.Message;
  }

  getStatus(): NodeStatus {
    return this.status;
  }
}

export type { DKGInfo, NodeProxyAddress };
export { InternalDKGInfo };
