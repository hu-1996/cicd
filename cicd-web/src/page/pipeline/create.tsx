import { useEffect, useState } from "react";
import { useNavigate, useSearchParams, Outlet } from "react-router-dom";
import type { TreeDataNode, TreeProps } from "antd";
import { Tree } from "antd";
import { DownOutlined } from "@ant-design/icons";
import { fetchRequest } from "../../utils/fetch";

export default function NewPipeline() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const [pipelineName, setPipelineName] = useState<string>("");

  const [treeData, setTreeData] = useState<TreeDataNode[]>([]);

  useEffect(() => {
    if (searchParams.get("id")) {
      loadPipelineDetail();
    } else {
      setTreeData([
        {
          title: "New Pipeline",
          key: "new_pipeline",
        },
      ]);
    }
  }, []);

  const loadPipelineDetail = async () => {
    const res = await fetchRequest("/api/pipeline/" + searchParams.get("id"), {
      method: "GET",
    });
    setPipelineName(res.name);
    setTreeData([
      {
        title: res.name,
        key: res.name,
        children: res.steps?.map((step: any) => {
          return {
            title: step.name,
            key: step.id,
          };
        }),
      },
    ]);
  };

  const onSelect: TreeProps["onSelect"] = (selectedKeys, info) => {
    console.log("selected", selectedKeys, info);
    if (selectedKeys[0] === "new_pipeline") {
      navigate("/new_pipeline/pipeline");
      return;
    } else if (selectedKeys[0] === pipelineName) {
      navigate("/new_pipeline/pipeline?id=" + searchParams.get("id"));
    } else {
      navigate(
        "/new_pipeline/step?id=" +
          searchParams.get("id") +
          "&step_id=" +
          selectedKeys[0]
      );
    }
  };

  return (
    <div className="flex justify-start">
      <Tree
        showLine
        switcherIcon={<DownOutlined />}
        onSelect={onSelect}
        defaultExpandAll
        treeData={treeData}
        className="w-[200px] p-5"
      />
      <Outlet />
    </div>
  );
}
