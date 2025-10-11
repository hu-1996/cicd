import { useEffect, useState } from "react";
import { useNavigate, useSearchParams, Outlet } from "react-router-dom";
import type { TreeDataNode, TreeProps } from "antd";
import { Tree } from "antd";
import { DownOutlined } from "@ant-design/icons";
import { fetchRequest } from "../../utils/fetch";

export default function NewPipeline() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

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
    setTreeData([
      {
        title: res.name,
        key: res.id,
        isLeaf: false,
        children: res.stages_and_steps?.map((item: any) => {
          return {
            title: item.name,
            key: item.id,
            isLeaf: item.type === "step",
            children:
              item.type === "stage"
                ? item.children.map((child: any) => {
                    return {
                      title: child.name,
                      key: child.id,
                      isLeaf: true,
                    };
                  })
                : undefined,
          };
        }),
      },
    ]);
  };

  const onSelect: TreeProps["onSelect"] = (selectedKeys, info) => {
    if (selectedKeys[0] === "new_pipeline") {
      navigate("/new_pipeline/pipeline");
      return;
    } else if (info.node.pos === '0-0' && selectedKeys[0] === Number(searchParams.get("id"))) {
      navigate("/new_pipeline/pipeline?id=" + searchParams.get("id"));
    } else {
      let navigateUri = `/new_pipeline/step?id=${searchParams.get(
        "id"
      )}&step_id=${selectedKeys[0]}`;
      if (!info.node.isLeaf) {
        navigateUri = `/new_pipeline/stage?id=${searchParams.get(
          "id"
        )}&stage_id=${selectedKeys[0]}`;
      }
      navigate(navigateUri);
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
    const sorts = data[0];
    let stages = [];
    let steps = [];
    for (let i = 0; i < (sorts?.children || []).length; i++) {
      const item = (sorts?.children || [])[i];
      if (item.isLeaf) {
        steps.push({
          id: item.key,
          sort: i,
        });
        for (let j = 0; j < (item.children || []).length; j++) {
          const step = (item.children || [])[j];
          steps.push({
            id: step.key,
            sort: j,
          });
        }
      } else {
        stages.push({
          id: item.key,
          sort: i,
        });
      }
    }
    await fetchRequest("/api/sort_stage_and_step/" + searchParams.get("id"), {
      method: "POST",
      body: JSON.stringify({
        stages,
        steps,
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
