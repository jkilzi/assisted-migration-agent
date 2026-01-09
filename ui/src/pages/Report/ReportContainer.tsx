import React, { useState } from "react";
import {
  Card,
  CardBody,
  MenuToggle,
  MenuToggleElement,
  Select,
  SelectList,
  SelectOption,
  Stack,
  StackItem,
  Tab,
  Tabs,
  TabTitleText,
} from "@patternfly/react-core";
import { useAppSelector } from "@shared/store";
import { Infra, InventoryData, VMs } from "@generated/index";
import Header from "./Header";
import Report from "./Report";
import { buildClusterViewModel, ClusterOption } from "./assessment-report/clusterView";

const ReportContainer: React.FC = () => {
  const { inventory } = useAppSelector((state) => state.collector);
  const [activeTab, setActiveTab] = useState<string | number>(0);
  const [selectedClusterId, setSelectedClusterId] = useState<string>("all");
  const [isClusterSelectOpen, setIsClusterSelectOpen] = useState(false);

  if (!inventory) {
    return null;
  }

  const infra = inventory?.vcenter?.infra as Infra | undefined;
  const vms = inventory?.vcenter?.vms as VMs | undefined;
  const clusters = inventory?.clusters as { [key: string]: InventoryData } | undefined;

  const clusterView = buildClusterViewModel({
    infra,
    vms,
    clusters,
    selectedClusterId,
  });

  const clusterSelectDisabled = clusterView.clusterOptions.length <= 1;

  const handleClusterSelect = (
    _event: React.MouseEvent<Element, MouseEvent> | undefined,
    value: string | number | undefined
  ): void => {
    if (typeof value === "string") {
      setSelectedClusterId(value);
    }
    setIsClusterSelectOpen(false);
  };

  return (
    <Stack hasGutter style={{ padding: "24px", width: "75%" }}>
      <Header>
        <Select
          isScrollable
          isOpen={isClusterSelectOpen}
          selected={clusterView.selectionId}
          onSelect={handleClusterSelect}
          onOpenChange={(isOpen: boolean) => {
            if (!clusterSelectDisabled) setIsClusterSelectOpen(isOpen);
          }}
          toggle={(toggleRef: React.Ref<MenuToggleElement>) => (
            <MenuToggle
              ref={toggleRef}
              isExpanded={isClusterSelectOpen}
              onClick={() => {
                if (!clusterSelectDisabled) {
                  setIsClusterSelectOpen((prev) => !prev);
                }
              }}
              isDisabled={clusterSelectDisabled}
              style={{ minWidth: "422px" }}
            >
              {clusterView.selectionLabel}
            </MenuToggle>
          )}
        >
          <SelectList>
            {clusterView.clusterOptions.map((option: ClusterOption) => (
              <SelectOption key={option.id} value={option.id}>
                {option.label}
              </SelectOption>
            ))}
          </SelectList>
        </Select>
      </Header>
      <StackItem>
        <Card>
          <CardBody>
            <Tabs
              activeKey={activeTab}
              onSelect={(_event, tabIndex) => setActiveTab(tabIndex)}
            >
              <Tab eventKey={0} title={<TabTitleText>Overview</TabTitleText>}>
                <div style={{ marginTop: 15 }}>
                  <Report clusterView={clusterView} />
                </div>
              </Tab>
              <Tab eventKey={1} title={<TabTitleText>Virtual Machines</TabTitleText>}>
                Virtual Machines content
              </Tab>
            </Tabs>
          </CardBody>
        </Card>
      </StackItem>
    </Stack>
  );
};

ReportContainer.displayName = "ReportContainer";

export default ReportContainer;
