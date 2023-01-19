import moment from "moment";
import {atom, atomFamily, selector, selectorFamily} from "recoil";
import {BuildAppUrl} from ".";

export type OrderData = Array<{date: Date, id: number, item: string, dateStr: string}>
export type InternalChartData = Map<string, Map<string, number>> 
function isOrderData(arg: any): arg is OrderData {
  let good = false
  if (arg && typeof(arg) == 'object') {
    if (arg && arg.length) {
      if (arg[0].id) {
        good = true
      }
    }
  }

  return good
}

export const createdArchiveDataId = atom<string>({
  key: "createdArchiveDataId",
})

export const archiveDataStore = atomFamily<OrderData, undefined>({
  key: "archiveDataStore",
  default: undefined,
  effects: [
    ({onSet}) => { 
      onSet(data => {
        fetch(BuildAppUrl("/api/archives"), {
          method: "POST",
          body: JSON.stringify(data),
          headers: {
            "Accept": "application/json",
            "Content-Type": "application/json",
          },
        })
      })
    },
  ]
})

export const neverMountedArchives = atom<boolean>({
  key: "neverMountedArchives",
  default: true,
})

export const neverMountedOrders = atom<boolean>({
  key: "neverMountedOrders",
  default: true,
})

export const last5kOrders = selector<{id: number, details: string}[]>({
  key: 'last5kOrders',
  get: async () => {
    const response = await fetch(BuildAppUrl("/api/orders"))
    if (response.ok) {
      return response.json()
    } else {
      return {error: 'failed to fetch orders'}
    }
  }
})

export const parsedLast5kOrders = selector<OrderData>({
  key: 'parsedLast5kOrders',
  get: async ({get}) => {
    const data: OrderData = []
    const v = get(last5kOrders)
    v.forEach(value => {
      data.push({
        date: moment(value.details.split("(")[0], "HH:mm:ss").toDate(),
        dateStr: buildDateStr(moment(value.details.split("(")[0], "HH:mm:ss").toDate()),
        item: randomItem(),
        id: value.id,
      })
    })

    return data
  }
})

export const getArchive = selectorFamily({
  key: 'archive',
  get: (archive: string | null) => async () => {
    if (archive === null) {
      return undefined
    }
    const response = await fetch(BuildAppUrl(`/api/archives?archiveUrl=${archive}`))
    if (response.ok) {
      const res = await response.text()
      let data: any
      try {
        data = JSON.parse(res)
      } catch (e) {
        console.log(`unexpected response from reporter service: ${JSON.stringify(res)}`) // debug
        return undefined
      }

      if (isOrderData(data)) {
        const orders: OrderData = []

        data.forEach((d: any) => {
          orders.push({
            date: moment(d.date).toDate(),
            dateStr: d.dateStr,
            id: d.id,
            item: d.item,
          })
        })

        const chartData = createProductCountChartData(orders)
        return {
          chartData,
          orders,
        }
      } else {
        console.log(`unexpected response from reporter service: ${JSON.stringify(data)}`) // debug
        return undefined
      }
    } else {
      return undefined
    }
  }
})

export const getArchives = selector<{url: string, name: string}[]>({
  key: 'archives',
  get: async () => {
    const retData: {name: string, url: string}[] = []
    const response = await fetch(BuildAppUrl("/api/archives"))
    if (response.ok) {
      const data: Array<string> = await response.json()
      data.forEach((d: any) => {
        retData.push({name: d.name, url: d.url})
      })
    }
    return retData.sort((a, b) => a.url > b.url ? -1 : 1)
  }
})

export const randomItem = (): string => {
  const objects = ["The Key", "The Lock", "The Fancy Lock", "The Safe"];
  return objects[Math.random() * objects.length >> 0]
}

const buildDateStr = (dateObj: Date): string => {
  var month = dateObj.getUTCMonth() + 1;
  var day = dateObj.getUTCDate();
  var year = dateObj.getUTCFullYear();
  return year + "/" + month + "/" + day;
}


const createProductCountChartData = (data: OrderData): Map<string, Map<string, number>> => {
    const data2: Map<string, OrderData> = new Map()
    data.forEach(d => {
      if (data2.has(d.item)) {
        const newData = data2.get(d.item)
        newData?.push(d)
      } else {
        data2.set(d.item, [d])
      }
    })

    const countData: Map<string, Map<string, number>> = new Map()
    for (const k of data2.keys()) {
      const data = data2.get(k)
      const res: Map<string, number> = new Map()
      data?.forEach(d => {
        if (res.has(d.dateStr)) {
          const v = res.get(d.dateStr)
          res.set(d.dateStr, v ? v + 1 : 1)
        } else {
          res.set(d.dateStr, 1)
        }
      })


      countData.set(k, res)
    }
    return countData
}

export const ordersChartData = selector<InternalChartData>({
  key: 'orderChartData',
  get: async ({get}) => {
    const data = get(parsedLast5kOrders)
    return createProductCountChartData(data)
  }
})
