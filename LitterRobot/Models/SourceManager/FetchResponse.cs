using System;

namespace LitterRobot.Models.SourceManager
{
    /// <summary>
    /// The response from the source
    /// </summary>
    public class FetchResponse
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string PowerStatus { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string SleepModeStartTime { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public DateTimeOffset LastSeen { get; set; } = DateTime.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string SleepModeEndTime { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool AutoOfflineDisabled { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public DateTime SetupDate { get; set; } = DateTime.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long DFICycleCount { get; set; } = long.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string CleanCycleWaitTimeMinutes { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string UnitStatus { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool IsOnboarded { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string DeviceType { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string LitterRobotNickname { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long CycleCount { get; set; } = long.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string PanelLockActive { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long CyclesAfterDrawerFull { get; set; } = long.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string LitterRobotSerial { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public long CycleCapacity { get; set; } = long.MinValue;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string LitterRobotId { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string NightLightActive { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool DidNotifyOffline { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string IsDFITriggered { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string SleepModeActive { get; set; } = string.Empty;

        /// <inheritdoc />
        public override string ToString() => $"LitterRobotID: {this.LitterRobotId}, Status: {this.UnitStatus}";
    }
}