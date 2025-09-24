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
        key: res.id,
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
    } else if (
      selectedKeys[0] === pipelineName ||
      selectedKeys.length === 0 ||
      (info.node.children ?? []).length > 0
    ) {
      navigate("/new_pipeline/pipeline?id=" + searchParams.get("id"));
    } else {
      console.log("selectedKeys[0]", selectedKeys);
      navigate(
        "/new_pipeline/step?id=" +
          searchParams.get("id") +
          "&step_id=" +
          selectedKeys[0]
      );
    }
  };

  const onDrop: TreeProps["onDrop"] = async (info) => {
    const dropKey = info.node.key;
    const dragKey = info.dragNode.key;
    const dropPos = info.node.pos.split("-");
    const dropPosition =
      info.dropPosition - Number(dropPos[dropPos.length - 1]); // the drop position relative to the drop node, inside 0, top -1, bottom 1

    const loop = (
      data: TreeDataNode[],
      key: React.Key,
      callback: (node: TreeDataNode, i: number, data: TreeDataNode[]) => void
    ) => {
      for (let i = 0; i < data.length; i++) {
        if (data[i].key === key) {
          return callback(data[i], i, data);
        }
        if (data[i].children) {
          loop(data[i].children!, key, callback);
        }
      }
    };
    const data = [...treeData];

    // Find dragObject
    let dragObj: TreeDataNode;
    loop(data, dragKey, (item, index, arr) => {
      arr.splice(index, 1);
      dragObj = item;
    });

    if (!info.dropToGap) {
      // Drop on the content
      loop(data, dropKey, (item) => {
        item.children = item.children || [];
        // where to insert. New item was inserted to the start of the array in this example, but can be anywhere
        item.children.unshift(dragObj);
      });
    } else {
      let ar: TreeDataNode[] = [];
      let i: number;
      loop(data, dropKey, (_item, index, arr) => {
        ar = arr;
        i = index;
      });
      if (dropPosition === -1) {
        // Drop on the top of the drop node
        ar.splice(i!, 0, dragObj!);
      } else {
        // Drop on the bottom of the drop node
        ar.splice(i! + 1, 0, dragObj!);
      }
    }
    setTreeData(data);
    await fetchRequest("/api/sort_step/" + searchParams.get("id"), {
      method: "POST",
      body: JSON.stringify({
        step_ids: data[0]?.children?.map((item) => item.key),
      }),
    });
  };

  return (
    <div className="flex justify-start">
      <Tree.DirectoryTree
        showLine
        draggable
        expandAction={false}
        switcherIcon={<DownOutlined />}
        onSelect={onSelect}
        treeData={treeData}
        className="w-[300px] p-5"
        onDrop={onDrop}
      />
      <Outlet />
    </div>
  );
}
