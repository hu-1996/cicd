import React, { useEffect } from "react";
import { Layout, Menu } from "antd";
import { Outlet, useNavigate } from "react-router-dom";
const { Header, Content } = Layout;

const App: React.FC = () => {
  const navigate = useNavigate();
  const handleMenuClick = (e: any) => {
    navigate(e.key);
  };

  useEffect(() => {
    if (location.pathname === "/") {
      redirectPipeline()
    }
  }, []);

  const redirectPipeline = () => {
    navigate("/pipeline");
  };

  return (
    <Layout>
      <Header
        style={{
          position: "sticky",
          top: 0,
          zIndex: 1,
          width: "100%",
          display: "flex",
          alignItems: "center",
        }}
      >
        <Menu
          theme="dark"
          mode="horizontal"
          defaultSelectedKeys={[window.location.pathname.split('/')[1]]}
          onClick={handleMenuClick}
          items={[
            {
              key: "pipeline",
              label: "pipeline",
            },
            {
              key: "runner",
              label: "执行机器",
            },
          ]}
          style={{ flex: 1, minWidth: 0 }}
        />
      </Header>
      <Content style={{ padding: "20px 20px", height: "calc(100vh - 64px)" }}>
        <Outlet />
      </Content>
    </Layout>
  );
};

export default App;
