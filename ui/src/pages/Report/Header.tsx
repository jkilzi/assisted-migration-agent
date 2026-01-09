import React, { ReactNode } from "react";
import { Card, CardBody, Content, Stack, StackItem, Title } from "@patternfly/react-core";

interface HeaderProps {
  children?: ReactNode;
}

const Header: React.FC<HeaderProps> = ({ children }) => {
  return (
    <StackItem>
      <Card>
        <CardBody>
          <Stack hasGutter>
            <StackItem>
              <Title headingLevel="h1" size="2xl">
                Migration Assessment Report
              </Title>
            </StackItem>
            <StackItem>
              <Content component="p">
                Presenting the information we were able to fetch from the discovery
                process
              </Content>
            </StackItem>
            {children && <StackItem>{children}</StackItem>}
          </Stack>
        </CardBody>
      </Card>
    </StackItem>
  );
};

Header.displayName = "Header";

export default Header;
