import {Bar} from "react-chartjs-2";
import './Orders.css';

import {Chart as ChartJS, ChartData, registerables} from 'chart.js';
import {useCallback, useEffect, useState} from 'react';
import {useRecoilCallback, useRecoilRefresher_UNSTABLE, useRecoilValue, useSetRecoilState} from 'recoil';
import {BuildAppUrl} from ".";
import {InternalChartData, last5kOrders, neverMountedOrders, OrderData, ordersChartData, parsedLast5kOrders} from './data';

ChartJS.register(...registerables)


type DisplayArchiveProps = {
  archiveId: string
}

export function DisplayCreatedLink({archiveId}: DisplayArchiveProps) {
  if (archiveId) {
    return <div className="displayLink">
      <span>
        <h5>Created Saved Report: {archiveId}!</h5>
      </span>
    </div>
  } else {
    return <div />
  }
}

type SaveOrderProps = {
  orders: OrderData
}


export function SaveOrder({orders}: SaveOrderProps) {
  const [archiveId, setArchiveId] = useState("")
  const [isSending, setIsSending] = useState(false)
  const sendRequest = useCallback(async (orders: OrderData) => {
    if (isSending) return
    setIsSending(true)
    const res = await fetch(BuildAppUrl("/api/archives"), {
      method: "POST",
      body: JSON.stringify(orders),
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
      },
    })
    const data = await res.text()
    if (data) {
      const jdata = JSON.parse(data)
      if (jdata && jdata.url) {
        setArchiveId(jdata.url)
      }
    }
    setIsSending(false)
  }, [isSending]) // update the callback if the state changes

  return (
    <div>
      <DisplayCreatedLink archiveId={archiveId} />
      <span className="saveReport">
        <button className="buttonClass" disabled={isSending || archiveId !== ""} onClick={() => sendRequest(orders)}>Save</button>
      </span>
    </div>
  )
}

export function OrderBody() {
  const orderCleanup = useRecoilRefresher_UNSTABLE(last5kOrders)
  const setMountedState = useSetRecoilState(neverMountedOrders)
  const neverMountedOrdersState = useRecoilCallback(({snapshot}) => async () => {
    const nm = await snapshot.getPromise(neverMountedOrders);
    return nm
  }, [])

  useEffect(() => {
    async function setMounted() {
      if (await neverMountedOrdersState()) {
        setMountedState(false)
      } else {
        orderCleanup()
      }

    }
    setMounted()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return <div>
    <Orders />
  </div>
}

export function Orders() {
  const chartData = useRecoilValue(ordersChartData)
  const orders = useRecoilValue(parsedLast5kOrders)

  return <div className="orders">
    <SaveOrder orders={orders} />
    <BarChart chartData={chartData} />
    <OrderListing orders={orders} />
  </div>
}


type OrderListingProps = {
  orders: OrderData | undefined
}

export const OrderListing = ({orders}: OrderListingProps) => {
  if (orders && orders?.length > 0) {
    return <table id="orders">
      <thead>
        <tr>
          <th>Transaction ID</th>
          <th>Date</th>
          <th>Item</th>
        </tr>
      </thead>
      <tbody>
        {orders.map(column =>
          <tr key={column.id}>
            <td className="td-data">{column.id}</td>
            <td className="td-data">{column.date.toLocaleString()}</td>
            <td className="td-data">{column.item}</td>
          </tr>
        )}
      </tbody>

    </table>

  } else {
    return <div>No data</div>
  }
}

type BarCharProps = {
  chartData: InternalChartData | undefined
}

export const BarChart = ({chartData}: BarCharProps) => {
  let data: ChartData<"bar", number[], string> = {
    labels: ["Products"],
    datasets: [],
  }
  if (chartData) {
    for (const k of chartData.keys()) {
      const newData = chartData.get(k)
      if (newData) {
        const counts: number[] = []
        for (const innerK of newData) {
          counts.push(innerK[1])
        }

        data.datasets.push({
          label: k,
          data: counts,
        })
      }
    }

    return (
      <div className="chart-container">
        <h2 style={{textAlign: "center"}}>Total Sales</h2>
        <Bar
          data={data}
          options={{
            plugins: {
              title: {
                display: true,
                text: "Sales distribution over the last 5k transactions"
              },
              legend: {
                display: true
              }
            }
          }}
        />
      </div>
    );
  } else {
    return <div className="chart-container"></div>
  }
};
