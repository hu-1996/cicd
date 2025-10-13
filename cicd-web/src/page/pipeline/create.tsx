import { useEffect, useState } from "react";
import { useNavigate, useSearchParams, Outlet } from "react-router-dom";
import type { TreeDataNode, TreeProps } from "antd";
import { Tree, Button } from "antd";
import {
  DownOutlined,
  FolderOpenOutlined,
  ArrowLeftOutlined,
} from "@ant-design/icons";
import { fetchRequest } from "../../utils/fetch";

export default function NewPipeline() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const [pipeline, setPipeline] = useState<any>({});
  const [treeData, setTreeData] = useState<TreeDataNode[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<any>([]);

  const pipelineId = searchParams.get("id");
  const search = searchParams.get("q");

  useEffect(() => {
    if (pipelineId) {
      loadPipelineDetail();
    } else {
      setPipeline({
        name: "New Pipeline",
      });
    }
  }, []);

  const loadPipelineDetail = async () => {
    const res = await fetchRequest("/api/pipeline/" + pipelineId, {
      method: "GET",
    });
    setPipeline(res);

    const treeData = res.stages_and_steps?.map((item: any) => {
      return {
        title: item.name,
        key: item.id,
        isLeaf: item.type === "step",
        children:
          item.type === "stage"
            ? item.children?.map((child: any) => {
                return {
                  title: child.name,
                  key: child.id,
                  isLeaf: true,
                };
              })
            : undefined,
      };
    });
    setTreeData(treeData);
  };

  const onPipeline = () => {
    setSelectedKeys([]);
    if (pipelineId) {
      navigate("/new_pipeline/pipeline?id=" + pipelineId);
    } else {
      navigate("/new_pipeline/pipeline");
    }
  };

  const onSelect: TreeProps["onSelect"] = (selectedKeys, info) => {
    setSelectedKeys(selectedKeys);
    let navigateUri = "";
    if (selectedKeys[0] === "new_pipeline") {
      navigateUri = "/new_pipeline/pipeline";
    } else if (
      info.node.pos === "0-0" &&
      selectedKeys[0] === Number(pipelineId)
    ) {
      navigateUri = "/new_pipeline/pipeline?id=" + pipelineId;
    } else {
      navigateUri = `/new_pipeline/step?id=${pipelineId}&step_id=${selectedKeys[0]}`;
      if (!info.node.isLeaf) {
        navigateUri = `/new_pipeline/stage?id=${pipelineId}&stage_id=${selectedKeys[0]}`;
      }
    }
    if (search) {
      navigateUri += "&q=" + search;
    }
    navigate(navigateUri);
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
    let stages = [];
    let steps = [];
    let sort = 0;
    for (let i = 0; i < data.length; i++) {
      const item = data[i];
      if (item.isLeaf) {
        steps.push({
          id: item.key,
          sort: sort++,
        });
      } else {
        stages.push({
          id: item.key,
          sort: sort++,
        });
        for (let j = 0; j < (item.children || []).length; j++) {
          const step = (item.children || [])[j];
          steps.push({
            id: step.key,
            stage_id: item.key,
            sort: sort++,
          });
        }
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
    <>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        className="text-left justify-start"
        onClick={() => {
          if (search) {
            navigate("/pipeline?q=" + search);
          } else {
            navigate("/pipeline");
          }
        }}
      >
        返回
      </Button>
      <div className="flex justify-start">
        <div className="bg-white">
          <div className="px-5 pt-5">
            <Button
              type={selectedKeys.length == 0 ? "primary" : "text"}
              icon={<FolderOpenOutlined />}
              block
              size="small"
              className="text-left justify-start p-[6px] mb-[4px]"
              onClick={onPipeline}
            >
              {pipeline.name}
            </Button>
          </div>
          <Tree.DirectoryTree
            showLine
            selectedKeys={selectedKeys}
            draggable={{ icon: false }}
            defaultExpandAll
            expandAction={false}
            switcherIcon={<DownOutlined />}
            onSelect={onSelect}
            treeData={treeData}
            className="w-[300px] px-5"
            onDrop={onDrop}
          />
        </div>
        <Outlet />
      </div>
    </>
  );
}
