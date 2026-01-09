import React from "react";

import { Infra, VMResourceBreakdown, VMs } from "@generated/index";
import { Bullseye, Content } from "@patternfly/react-core";

import { ClusterViewModel } from "./assessment-report/clusterView";
import { Dashboard } from "./assessment-report/Dashboard";

interface ReportProps {
  clusterView: ClusterViewModel;
}

const Report: React.FC<ReportProps> = ({ clusterView }) => {
  const hasClusterScopedData =
    Boolean(clusterView.viewInfra) &&
    Boolean(clusterView.viewVms) &&
    Boolean(clusterView.cpuCores) &&
    Boolean(clusterView.ramGB);

  return hasClusterScopedData ? (
    <Dashboard
      infra={clusterView.viewInfra as Infra}
      vms={clusterView.viewVms as VMs}
      cpuCores={clusterView.cpuCores as VMResourceBreakdown}
      ramGB={clusterView.ramGB as VMResourceBreakdown}
      clusters={clusterView.viewClusters}
      isAggregateView={clusterView.isAggregateView}
      clusterFound={clusterView.clusterFound}
    />
  ) : (
    <Bullseye style={{ width: "100%" }}>
      <Content>
        <Content component="p">
          {clusterView.isAggregateView
            ? "This assessment does not have report data yet."
            : "No data is available for the selected cluster."}
        </Content>
      </Content>
    </Bullseye>
  );
};

Report.displayName = "Report";

export default Report;
