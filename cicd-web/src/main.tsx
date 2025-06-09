import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { ConfigProvider } from "antd";
import Pipeline from "./page/pipeline";
import Setting from "./page/pipeline/setting.tsx";
import NewPipeline from "./page/pipeline/create.tsx";
import CreatePipeline from "./page/pipeline/create_pipeline.tsx";
import CreateStep from "./page/pipeline/create_step.tsx";
import History from "./page/pipeline/history.tsx";
import Logs from "./page/pipeline/logs.tsx";
import Runner from "./page/runner/index.tsx";
import Layout from "./layout.tsx";
import Login from "./page/login.tsx";
import "antd/dist/reset.css";
import "./index.css";

createRoot(document.getElementById("root")!).render(
  <ConfigProvider>
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Layout />}>
          <Route path="/pipeline" element={<Pipeline />} />
          <Route path="/setting" element={<Setting />} />
          <Route path="/new_pipeline" element={<NewPipeline />}>
            <Route path="pipeline" element={<CreatePipeline />} />
            <Route path="step" element={<CreateStep />} />
          </Route>
          <Route path="/history" element={<History />} />
          <Route path="/logs" element={<Logs />} />
          <Route path="/runner" element={<Runner />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </ConfigProvider>
);
