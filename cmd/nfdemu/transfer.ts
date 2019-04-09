import * as loglevel from "loglevel";
import ndn = require("ndn-js");
import * as net from "net";

import { AppConn } from "./app-conn";
import { FwConn } from "./fw-conn";
import { Packet, PktType } from "./packet";
import { PendingInterest } from "./pending-interest";
import { PendingInterestList } from "./pending-interest-list";

const ndnjs = ndn as any;

const keyChain = new ndn.KeyChain("pib-memory:", "tpm-memory:");
const signingInfo = new ndnjs.SigningInfo(ndnjs.SigningInfo.SignerType.SHA256);

export class Transfer {
  private fc: FwConn;
  private ac: AppConn;
  private pil: PendingInterestList;
  private log: loglevel.Logger;

  constructor(appSocket: net.Socket) {
    this.pil = new PendingInterestList();
    this.fc = new FwConn();
    this.ac = new AppConn(appSocket);
    this.log = loglevel;
  }

  public begin(): void {
    this.fc.on("error", (reason) => { this.log.error(reason); });
    this.fc.on("connected", () => {
      this.log = loglevel.getLogger("" + this.fc.faceId);
      this.log.info("> CONNECTED", "pid=" + process.pid);
      this.ac.begin();
    });
    this.fc.on("packet", (pkt: Packet) => { this.handleFwPacket(pkt); });

    this.ac.on("close", () => {
      this.log.info("< CLOSE");
      this.fc.close();
    });
    this.ac.on("packet", (pkt: Packet) => { this.handleAppPacket(pkt); });
  }

  private handleFwPacket(pkt: Packet): void {
    this.log.debug(">" + pkt.toString(), pkt.pitToken || "no-token");
    if (pkt.type === PktType.Interest) {
      this.pil.insert(new PendingInterest(pkt.interest!, pkt.pitToken));
    }
    this.ac.send(pkt.wireEncode(false, false));
  }

  private handleAppPacket(pkt: Packet): void {
    switch (pkt.type) {
      case PktType.Interest:
        const prefixRegName = pkt.tryParsePrefixReg();
        if (prefixRegName) {
          this.handleAppPrefixReg(pkt.interest!, prefixRegName);
          return;
        }
        break;
      case PktType.Data:
      case PktType.Nack:
        const pi = this.pil.find(pkt.name!);
        if (pi) {
          pkt.pitToken = pi.pitToken;
        }
        break;
    }
    this.log.debug("<" + pkt.toString(), pkt.pitToken || "no-token");
    this.fc.send(pkt.wireEncode(true, true));
  }

  private handleAppPrefixReg(interest: ndn.Interest, name: ndn.Name): void {
    const respond = (cr: any) => {
      const data = new ndn.Data();
      data.setName(interest.getName());
      data.setContent(cr.wireEncode());
      keyChain.sign(data, signingInfo);
      this.ac.send(data.wireEncode().buf());
    };

    this.log.info("<R", name.toUri());
    this.fc.registerPrefix(name)
    .then(() => {
      this.log.info(">R", name.toUri());
      const flags = new ndnjs.ForwardingFlags();
      flags.setChildInherit(false);
      flags.setCapture(true);
      const cp = new ndnjs.ControlParameters();
      cp.setName(name);
      cp.setFaceId(1);
      cp.setOrigin(0);
      cp.setCost(0);
      cp.setForwardingFlags(flags);
      const cr = new ndnjs.ControlResponse();
      cr.setStatusCode(200).setStatusText("OK");
      cr.setBodyAsControlParameters(cp);
      respond(cr);
    })
    .catch((reason) => {
      this.log.error(">R", name.toUri(), reason);
      const cr = new ndnjs.ControlResponse();
      cr.setStatusCode(500).setStatusText("ERROR");
      respond(cr);
    });
  }
}
